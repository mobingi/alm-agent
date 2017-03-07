package cmd

import (
	"github.com/mobingilabs/go-modaemon/api"
	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/container"
	"github.com/urfave/cli"
)

func Stop(c *cli.Context) error {
	conf, err := config.LoadFromFile(c.String("config"))
	if err != nil {
		return err
	}

	apiClient, err := api.NewClient(conf)
	if err != nil {
		return err
	}

	s, err := apiClient.GetServerConfig(c.String("serverconfig"))
	if err != nil {
		return err
	}

	d, err := container.NewDocker(s)
	if err != nil {
		return err
	}

	activeContainer, err := d.GetContainer("active")
	d.MapPort(activeContainer) // For regenerating port map information
	d.UnmapPort()
	d.StopContainer(activeContainer)
	d.RemoveContainer(activeContainer)

	return nil
}
