package cmd

import (
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/mobingi/alm-agent/api"
	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/container"
	"github.com/mobingi/alm-agent/util"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

// Stop alm-agent start
func Stop(c *cli.Context) error {
	util.AgentID()
	conf, err := config.LoadFromFile(c.String("config"))
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	api.SetConfig(conf)

	err = api.GetAccessToken()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	s, err := api.GetServerConfig(c.String("serverconfig"))
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	d, err := container.NewDocker(conf, s)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	activeContainer, err := d.GetContainer("active")
	d.MapPort(activeContainer) // For regenerating port map information
	d.UnmapPort()

	dockerCli, err := client.NewEnvClient()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	containers, err := dockerCli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	timeout := 3 * time.Second
	for _, c := range containers {
		if err := dockerCli.ContainerStop(context.Background(), c.ID, &timeout); err != nil {
			return cli.NewExitError(err, 1)
		}
		if err := dockerCli.ContainerRemove(context.Background(), c.ID, types.ContainerRemoveOptions{}); err != nil {
			return cli.NewExitError(err, 1)
		}
	}

	return nil
}
