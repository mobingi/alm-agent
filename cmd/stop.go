package cmd

import (
	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/container"
	"github.com/mobingilabs/go-modaemon/server_config"
	"github.com/urfave/cli"
)

func Stop(c *cli.Context) error {
	conf, err := config.LoadFromFile(c.String("config"))
	if err != nil {
		return err
	}

	s, err := serverConfig.Get(conf.ServerConfigAPIEndPoint)
	if err != nil {
		return err
	}

	d, err := container.NewDocker(s)
	if err != nil {
		return err
	}

	activeContainer, err := d.GetContainer("active")
	d.UnmapPort(activeContainer)
	d.StopContainer(activeContainer)
	d.RemoveContainer(activeContainer)

	return nil
}
