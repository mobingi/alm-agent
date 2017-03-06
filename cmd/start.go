package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mobingilabs/go-modaemon/api"
	"github.com/mobingilabs/go-modaemon/code"
	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/container"
	"github.com/urfave/cli"
)

func Start(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		log.SetLevel(log.DebugLevel)
		log.Debug("Loglevel is set to DebugLevel.")
	}

	conf, err := config.LoadFromFile(c.String("config"))
	if err != nil {
		return err
	}

	apiClient, err := api.NewClient(conf)
	if err != nil {
		return err
	}

	s, err := apiClient.GetServerConfig()
	if err != nil {
		return err
	}

	codeDir := ""
	if s.Code != "" {
		code := code.New(s)
		codeDir, err = code.Get()
		if err != nil {
			return err
		}
	}

	d, err := container.NewDocker(s)
	if err != nil {
		return err
	}

	newContainer, err := d.StartContainer("active", codeDir)
	if err != nil {
		return err
	}

	err = d.MapPort(newContainer)
	if err != nil {
		return err
	}

	return nil
}
