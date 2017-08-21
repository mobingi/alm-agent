package cmd

import (
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/mobingi/alm-agent/api"
	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/container"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
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

	d, err := container.NewDocker(conf, s)
	if err != nil {
		return err
	}

	activeContainer, err := d.GetContainer("active")
	d.MapPort(activeContainer) // For regenerating port map information
	d.UnmapPort()

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return err
	}

	timeout := 3 * time.Second
	for _, c := range containers {
		if err := cli.ContainerStop(context.Background(), c.ID, &timeout); err != nil {
			return err
		}
		if err := cli.ContainerRemove(context.Background(), c.ID, types.ContainerRemoveOptions{}); err != nil {
			return err
		}
	}

	return nil
}
