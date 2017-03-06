package main

import (
	"os"
	"sort"

	"github.com/mobingilabs/go-modaemon/cmd"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "go-modaemon"
	app.Version = "0.1.1.dev"
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
	}

	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "start active container",
			Action: cmd.Start,
			Flags:  flags,
		},
		{
			Name:   "stop",
			Usage:  "stop active container",
			Action: cmd.Stop,
			Flags:  flags,
		},
		{
			Name:   "update",
			Usage:  "update code and image, then switch container",
			Action: cmd.Update,
			Flags:  flags,
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	app.Run(os.Args)
}
