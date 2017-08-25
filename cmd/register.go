package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

// Register alm-agent register
func Register(c *cli.Context) error {
	log.Warn("TODO: mkdirs")
	log.Warn("TODO: create known_hosts")
	log.Warn("TODO: crontab entry")
	err := Ensure(c)
	return err
}
