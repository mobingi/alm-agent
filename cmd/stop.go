package cmd

import (
	"github.com/mobingilabs/go-modaemon/api"
	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/container"
	molog "github.com/mobingilabs/go-modaemon/log"
	"github.com/mobingilabs/go-modaemon/util"
	"github.com/urfave/cli"
)

func Stop(c *cli.Context) error {
	serverid, err := util.GetServerID()
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

	d, err := container.NewDocker(conf, s)
	if err != nil {
		return err
	}

	activeContainer, err := d.GetContainer("active")
	d.MapPort(activeContainer) // For regenerating port map information
	d.UnmapPort()
	d.StopContainer(activeContainer)
	d.RemoveContainer(activeContainer)

	ld, err := molog.NewDocker(conf, serverid)
	if err != nil {
		return err
	}

	logContainer, err := ld.GetContainer("mo-awslogs")
	ld.StopContainer(logContainer)
	ld.RemoveContainer(logContainer)

	return nil
}
