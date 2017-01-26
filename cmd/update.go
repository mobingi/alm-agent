package cmd

import (
	"github.com/mobingilabs/go-modaemon/api"
	"github.com/mobingilabs/go-modaemon/code"
	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/container"
	"github.com/urfave/cli"
)

func Update(c *cli.Context) error {
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

	code := code.New(s)

	codeUpdated, err := code.CheckUpdate()
	if err != nil {
		return err
	}

	var dir string
	if codeUpdated {
		dir, err = code.Get()
		if err != nil {
			return err
		}
	} else {
		dir = code.Path
	}

	d, err := container.NewDocker(s)
	if err != nil {
		return err
	}

	imageUpdated, err := d.CheckImageUpdated()
	if err != nil {
		return err
	}

	oldContainer, err := d.GetContainer("active")
	if err != nil {
		return err
	}

	if oldContainer == nil {
		return Start(c)
	}

	if !codeUpdated && !imageUpdated {
		return nil
	}

	d.MapPort(oldContainer) // For regenerating port map information

	newContainer, err := d.StartContainer("standby", dir)
	if err != nil {
		return err
	}

	d.UnmapPort()
	d.MapPort(newContainer)

	d.StopContainer(oldContainer)
	d.RemoveContainer(oldContainer)

	d.RenameContainer(newContainer, "active")

	return nil
}
