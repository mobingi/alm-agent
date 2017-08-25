package cmd

import "github.com/urfave/cli"

// Register alm-agent register
func Register(c *cli.Context) error {
	// TODO: mkdirs
	// TODO: create known_hosts
	// TODO: crontab entry
	err := Ensure(c)
	return err
}
