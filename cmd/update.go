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

	stsToken, err := apiClient.GetStsToken()
	if err != nil {
		return err
	}

	apiClient.WriteTempToken(stsToken)

	s, err := apiClient.GetServerConfig(c.String("serverconfig"))
	if err != nil {
		return err
	}

	codeDir := ""
	codeUpdated := false
	if s.Code != "" {
		code := code.New(s)

		codeUpdated, err = code.CheckUpdate()
		if err != nil {
			return err
		}

		if codeUpdated {
			codeDir, err = code.Get()
			if err != nil {
				return err
			}
		} else {
			codeDir = code.Path
		}
	}

	d, err := container.NewDocker(conf, s)
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

	newContainer, err := d.StartContainer("standby", codeDir)
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
