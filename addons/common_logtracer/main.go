package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	docker "github.com/docker/docker"
	"github.com/docker/docker/api/types"
)

var logDirPath = "/var/log/alm-agent/container/log/"

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Please pass container id as 1st arg, bye.")
		os.Exit(int(0))
	} else if len(os.Args) == 2 {
		// exec itself
		cmd := exec.Command(os.Args[0], "--child", os.Args[1])
		cmd.Start()
	} else {
		// run like daemon
		var ctid = os.Args[2]
		ctxWithTO, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cli, err := docker.NewEnvClient()
		if err != nil {
			panic(err)
		}

		// check existance
		c, err := cli.ContainerInspect(ctxWithTO, ctid)
		if err != nil {
			fmt.Printf("%s\n is not found. bye!", ctid)
			os.Exit(int(1))
		}

		ctx := context.Background()

		options := types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Tail:       "all",
		}
		responseBody, err := cli.ContainerLogs(ctx, ctid, options)
		defer responseBody.Close()

		go func() {
			logfile := logDirPath + "dockerrun.active.log"
			dst, err := os.Create(logfile)
			if err != nil {
				panic(err)
			}
			defer dst.Close()
			if c.Config.Tty {
				io.Copy(dst, responseBody)
			} else {
				StdCopy(dst, dst, responseBody)
			}
		}()

		ctxBG := context.Background()
		wch, errC := cli.ContainerWait(ctxBG, ctid, "")

		go func() {
			for {
				a := <-wch
				os.Exit(int(a.StatusCode))
			}
		}()

		if err := <-errC; err != nil {
			log.Fatal(err)
		}
	}
	os.Exit(0)
}

// docker-ce/components/engine/pkg/stdcopy/stdcopy.go

// StdType is the type of standard stream
// a writer can multiplex to.
type StdType byte

const (
	// Stdin represents standard input stream type.
	Stdin StdType = iota
	// Stdout represents standard output stream type.
	Stdout
	// Stderr represents standard error steam type.
	Stderr
	// Systemerr represents errors originating from the system that make it
	// into the the multiplexed stream.
	Systemerr

	stdWriterPrefixLen = 8
	stdWriterFdIndex   = 0
	stdWriterSizeIndex = 4

	startingBufLen = 32*1024 + stdWriterPrefixLen + 1
)

var bufPool = &sync.Pool{New: func() interface{} { return bytes.NewBuffer(nil) }}

// stdWriter is wrapper of io.Writer with extra customized info.
type stdWriter struct {
	io.Writer
	prefix byte
}

// Write sends the buffer to the underneath writer.
// It inserts the prefix header before the buffer,
// so stdcopy.StdCopy knows where to multiplex the output.
// It makes stdWriter to implement io.Writer.
func (w *stdWriter) Write(p []byte) (n int, err error) {
	if w == nil || w.Writer == nil {
		return 0, errors.New("Writer not instantiated")
	}
	if p == nil {
		return 0, nil
	}

	header := [stdWriterPrefixLen]byte{stdWriterFdIndex: w.prefix}
	binary.BigEndian.PutUint32(header[stdWriterSizeIndex:], uint32(len(p)))
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Write(header[:])
	buf.Write(p)

	n, err = w.Writer.Write(buf.Bytes())
	n -= stdWriterPrefixLen
	if n < 0 {
		n = 0
	}

	buf.Reset()
	bufPool.Put(buf)
	return
}

// NewStdWriter instantiates a new Writer.
// Everything written to it will be encapsulated using a custom format,
// and written to the underlying `w` stream.
// This allows multiple write streams (e.g. stdout and stderr) to be muxed into a single connection.
// `t` indicates the id of the stream to encapsulate.
// It can be stdcopy.Stdin, stdcopy.Stdout, stdcopy.Stderr.
func NewStdWriter(w io.Writer, t StdType) io.Writer {
	return &stdWriter{
		Writer: w,
		prefix: byte(t),
	}
}

// StdCopy is a modified version of io.Copy.
//
// StdCopy will demultiplex `src`, assuming that it contains two streams,
// previously multiplexed together using a StdWriter instance.
// As it reads from `src`, StdCopy will write to `dstout` and `dsterr`.
//
// StdCopy will read until it hits EOF on `src`. It will then return a nil error.
// In other words: if `err` is non nil, it indicates a real underlying error.
//
// `written` will hold the total number of bytes written to `dstout` and `dsterr`.
func StdCopy(dstout, dsterr io.Writer, src io.Reader) (written int64, err error) {
	var (
		buf       = make([]byte, startingBufLen)
		bufLen    = len(buf)
		nr, nw    int
		er, ew    error
		out       io.Writer
		frameSize int
	)

	for {
		// Make sure we have at least a full header
		for nr < stdWriterPrefixLen {
			var nr2 int
			nr2, er = src.Read(buf[nr:])
			nr += nr2
			if er == io.EOF {
				if nr < stdWriterPrefixLen {
					return written, nil
				}
				break
			}
			if er != nil {
				return 0, er
			}
		}

		stream := StdType(buf[stdWriterFdIndex])
		// Check the first byte to know where to write
		switch stream {
		case Stdin:
			fallthrough
		case Stdout:
			// Write on stdout
			out = dstout
		case Stderr:
			// Write on stderr
			out = dsterr
		case Systemerr:
			// If we're on Systemerr, we won't write anywhere.
			// NB: if this code changes later, make sure you don't try to write
			// to outstream if Systemerr is the stream
			out = nil
		default:
			return 0, fmt.Errorf("Unrecognized input header: %d", buf[stdWriterFdIndex])
		}

		// Retrieve the size of the frame
		frameSize = int(binary.BigEndian.Uint32(buf[stdWriterSizeIndex : stdWriterSizeIndex+4]))

		// Check if the buffer is big enough to read the frame.
		// Extend it if necessary.
		if frameSize+stdWriterPrefixLen > bufLen {
			buf = append(buf, make([]byte, frameSize+stdWriterPrefixLen-bufLen+1)...)
			bufLen = len(buf)
		}

		// While the amount of bytes read is less than the size of the frame + header, we keep reading
		for nr < frameSize+stdWriterPrefixLen {
			var nr2 int
			nr2, er = src.Read(buf[nr:])
			nr += nr2
			if er == io.EOF {
				if nr < frameSize+stdWriterPrefixLen {
					return written, nil
				}
				break
			}
			if er != nil {
				return 0, er
			}
		}

		// we might have an error from the source mixed up in our multiplexed
		// stream. if we do, return it.
		if stream == Systemerr {
			return written, fmt.Errorf("error from daemon in stream: %s", string(buf[stdWriterPrefixLen:frameSize+stdWriterPrefixLen]))
		}

		// Write the retrieved frame (without header)
		nw, ew = out.Write(buf[stdWriterPrefixLen : frameSize+stdWriterPrefixLen])
		if ew != nil {
			return 0, ew
		}

		// If the frame has not been fully written: error
		if nw != frameSize {
			return 0, io.ErrShortWrite
		}
		written += int64(nw)

		// Move the rest of the buffer to the beginning
		copy(buf, buf[frameSize+stdWriterPrefixLen:])
		// Move the index
		nr -= frameSize + stdWriterPrefixLen
	}
}
