package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/mobingi/alm-agent/cmd"
	"github.com/mobingi/alm-agent/versions"
	"github.com/urfave/cli"
)

var agentConfigPath = "/opt/mobingi/etc/alm-agent.cfg"

func init() {
	log.SetOutput(os.Stdout)
	cli.ErrWriter = &FatalWriter{cli.ErrWriter}
}

func globalOptions(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		log.SetLevel(log.DebugLevel)
		log.Debug("Loglevel is set to DebugLevel.")
	}

	return nil
}

func beforeActions(c *cli.Context) error {
	globalOptions(c)
	if c.GlobalBool("autoupdate") {
		versions.AutoUpdate(golatest())
	}
	log.Debugf("Set provider to %#v", c.GlobalString("provider"))
	return nil
}

// ReleaseJSONURL builds URL of go-latest json
func ReleaseJSONURL() string {
	return strings.Join([]string{versions.URLBase, versions.Branch, "/current/version_info.json"}, "")
}

func golatest() *versions.GoLatest {
	v := &versions.GoLatest{}
	v.Version = versions.Version
	v.Message = versions.Revision
	v.URL = ReleaseJSONURL()
	return v
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		b, _ := json.MarshalIndent(golatest(), "", "  ")
		fmt.Println(string(b))
	}

	app := cli.NewApp()
	app.Name = "alm-agent"
	app.Version = versions.Version
	app.Usage = ""

	// Gloabl Flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "show debug logs",
		},
		cli.BoolFlag{
			Name:  "autoupdate, U",
			Usage: "auto update before run",
		},
		cli.StringFlag{
			Name:  "provider, P",
			Value: "aws",
			Usage: "set `Provider`",
		},
	}

	// Common Flags for commands
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: agentConfigPath,
			Usage: "Load configuration from `FILE`",
		},
		cli.StringFlag{
			Name:  "serverconfig, sc",
			Usage: "Load ServerConfig from `URL`. ask to API by default",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "register",
			Usage:  "initialize alm-agent and start containers",
			Action: cmd.Register,
			Flags:  flags,
			Before: beforeActions,
		},
		{
			Name:   "ensure",
			Usage:  "start or update containers",
			Action: cmd.Ensure,
			Flags:  flags,
			Before: beforeActions,
		},
		{
			Name:   "stop",
			Usage:  "stop active container",
			Action: cmd.Stop,
			Flags:  flags,
			Before: globalOptions,
		},
		{
			Name:   "noop",
			Usage:  "run without container actions.",
			Action: func(c *cli.Context) error { return nil },
			Flags:  flags,
			Before: beforeActions,
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	app.Run(os.Args)
}
