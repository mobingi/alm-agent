package cmd

import (
	"github.com/mobingilabs/go-modaemon/code"
	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/container"
	"github.com/mobingilabs/go-modaemon/server_config"
	"github.com/urfave/cli"
)

func Start(c *cli.Context) error {
	conf, err := config.LoadFromFile(c.String("config"))
	if err != nil {
		return err
	}

	s, err := serverConfig.Get(conf.ServerConfigAPIEndPoint)
	if err != nil {
		return err
	}

	code := code.Code{
		URL: s.Code,
		Ref: s.GitReference,
	}

	dir, err := code.Get()
	if err != nil {
		return err
	}

	d, err := container.NewDocker(s)
	if err != nil {
		return err
	}

	newContainer, err := d.StartContainer("active", dir)
	if err != nil {
		return err
	}

	err = d.MapPort(newContainer)
	if err != nil {
		return err
	}

	return nil
}
