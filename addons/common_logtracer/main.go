package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	docker "docker.io/go-docker"
	"docker.io/go-docker/api/types"
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
		_, err = cli.ContainerInspect(ctxWithTO, ctid)
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

		go func() {
			logfile := logDirPath + fmt.Sprintf("dockerrun.%s.log", ctid)
			dst, err := os.Create(logfile)
			if err != nil {
				panic(err)
			}
			defer dst.Close()
			io.Copy(dst, responseBody)
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
