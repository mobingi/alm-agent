package cmd

import (
	"github.com/mobingilabs/go-modaemon/code"
	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/container"
	"github.com/mobingilabs/go-modaemon/server_config"
	"github.com/urfave/cli"
)

func Update(c *cli.Context) error {
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
	}

	d, err := container.NewDocker(s)
	if err != nil {
		return err
	}

	oldContainer, err := d.GetContainer("active")
	d.MapPort(oldContainer) // For regenerating port map information

	newContainer, err := d.StartContainer("standby", dir)

	d.UnmapPort()
	d.MapPort(newContainer)

	d.StopContainer(oldContainer)
	d.RemoveContainer(oldContainer)

	d.RenameContainer(newContainer, "active")

	return nil
}
