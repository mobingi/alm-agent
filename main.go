package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	_ "github.com/mobingi/alm-agent/statik"
	log "github.com/sirupsen/logrus"

	"github.com/mobingi/alm-agent/cmd"
	"github.com/mobingi/alm-agent/metavars"
	"github.com/mobingi/alm-agent/util"
	"github.com/mobingi/alm-agent/versions"
	"github.com/stvp/rollbar"
	"github.com/urfave/cli"
)

var agentConfigPath = "/opt/mobingi/etc/alm-agent.cfg"

// RollbarToken is post_client_item token of rollbar.com
// it should be set on build.
var RollbarToken string

func init() {
	log.SetOutput(os.Stdout)
	cli.ErrWriter = &FatalWriter{cli.ErrWriter}
}

// FatalWriter uses for cliErrWriter
type FatalWriter struct {
	cliErrWriter io.Writer
}

func (f *FatalWriter) Write(p []byte) (n int, err error) {
	log.Error(string(p))
	return 0, nil
}

func globalOptions(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		log.SetLevel(log.DebugLevel)
		log.Debug("Loglevel is set to DebugLevel.")
	}

	if c.GlobalBool("disablereport") || versions.Revision == "local-build" {
		metavars.ReportDisabled = true
	}

	// initialize or load AgentID
	util.AgentID()

	// initialize rollbar client
	rollbar.Token = RollbarToken
	rollbar.Environment = metavars.AgentID
	rollbar.Platform = "client"

	return nil
}

func beforeActions(c *cli.Context) error {
	globalOptions(c)
	if c.GlobalBool("autoupdate") {
		versions.AutoUpdate(golatest())
	}
	log.Debugf("Set provider to %#v", c.GlobalString("provider"))

	if c.Command.Name == "ensure" {
		if c.Bool("immediately") {
			log.Debug("Splay skipped due to immediately flag was set")
			return nil
		}
		rand.Seed(int64(os.Getpid()))
		splay := rand.Intn(30000)
		log.Debugf("Wait %d milliseconds...", splay)
		time.Sleep(time.Duration(splay) * time.Millisecond)
	}
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
	// report panic to rollbar
	defer func() {
		rec := recover()
		if rec != nil {
			log.Errorf("Agent Crashed!: %s", rec)
			debug.PrintStack()
			if !metavars.ReportDisabled {
				rollbar.Error(rollbar.ERR, fmt.Errorf("%s", rec))
				rollbar.Wait()
			}
			os.Exit(1)
		}
	}()

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
		cli.BoolFlag{
			Name:  "disablereport, N",
			Usage: "Do not send crash report to rollbar.",
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
		cli.BoolFlag{
			Name:  "immediately, I",
			Usage: "It skips sleep to run ensure immediately(only effects for ensure). false by default.",
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

	defer util.UnLock("alm-agent")
	if err := util.Lock("alm-agent"); err != nil {
		log.Info("Other alm-agent running... Exit.")
		os.Exit(0)
	}

	app.Run(os.Args)
	rollbar.Wait()
}
