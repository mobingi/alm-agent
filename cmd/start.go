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
	log.Debug("Step: config.LoadFromFile")
	conf, err := config.LoadFromFile(c.String("config"))
	if err != nil {
		return err
	}
	log.Debugf("%#v", conf)

	log.Debug("Step: api.NewClient")
	apiClient, err := api.NewClient(conf)
	if err != nil {
		return err
	}
	log.Debugf("%#v", apiClient)

	log.Debug("Step: apiClient.GetServerConfig")
	log.Debugf("Flag: %#v", c.String("serverconfig"))
	s, err := apiClient.GetServerConfig(c.String("serverconfig"))
	if err != nil {
		return err
	}
	log.Debugf("%#v", s)

	codeDir := ""
	if s.Code != "" {
		code := code.New(s)
		codeDir, err = code.Get()
		if err != nil {
			return err
		}
	}

	log.Debug("Step: container.NewDocker")
	d, err := container.NewDocker(s)
	if err != nil {
		return err
	}
	log.Debugf("%#v", d)

	log.Debug("Step: d.StartContainer")
	newContainer, err := d.StartContainer("active", codeDir)
	if err != nil {
		return err
	}
	log.Debugf("%#v", newContainer)

	log.Debug("Step: d.MapPort")
	err = d.MapPort(newContainer)
	if err != nil {
		return err
	}

	return nil
}
