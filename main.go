package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/mobingilabs/go-modaemon/cmd"
	"github.com/urfave/cli"
)

func init() {
	cli.ErrWriter = &FatalWriter{cli.ErrWriter}
}

func globalOptions(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		log.SetLevel(log.DebugLevel)
		log.Debug("Loglevel is set to DebugLevel.")
	}

	return nil
}

// could be overwritten by LDFLAGS e.g) 'main.version=$(VERSION)'
type golatest struct {
	Version string `json:"version"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

var (
	version  = "0.1.1-dev"
	revision = "local-build"
	urlBase  = "https://download.labs.mobingi.com/go-modaemon/"
	branch   = "develop"
	binVer   = "current"
)

// ReleaseJSONURL builds URL of go-latest json
func ReleaseJSONURL() string {
	return strings.Join([]string{urlBase, branch, "/current/version_info.json"}, "")
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		golatest := &golatest{}
		golatest.Version = version
		golatest.Message = revision
		golatest.URL = ReleaseJSONURL()
		b, _ := json.MarshalIndent(golatest, "", "  ")
		fmt.Println(string(b))
	}

	app := cli.NewApp()
	app.Name = "go-modaemon"
	app.Version = version
	app.Usage = ""

	// Gloabl Flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "show debug logs",
		},
	}

	// Common Flags for commands
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "/opt/modaemon/modaemon.cfg",
			Usage: "Load configuration from `FILE`",
		},
		cli.StringFlag{
			Name:  "serverconfig, sc",
			Usage: "Load ServerConfig from `URL`. ask to API by default",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "start active container",
			Action: cmd.Start,
			Flags:  flags,
			Before: globalOptions,
		},
		{
			Name:   "stop",
			Usage:  "stop active container",
			Action: cmd.Stop,
			Flags:  flags,
			Before: globalOptions,
		},
		{
			Name:   "update",
			Usage:  "update code and image, then switch container",
			Action: cmd.Update,
			Flags:  flags,
			Before: globalOptions,
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	app.Run(os.Args)
}

type FatalWriter struct {
	cliErrWriter io.Writer
}

func (f *FatalWriter) Write(p []byte) (n int, err error) {
	log.Error(string(p))
	return 0, nil
}
