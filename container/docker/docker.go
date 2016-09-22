package docker

import (
	"bytes"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

func newClient() (*client.Client, error) {
	defaultHeaders := map[string]string{"User-Agent": "modaemon"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.24", nil, defaultHeaders)

	return cli, err
}

func ImagePull(ref string) error {
	cli, err := newClient()
	if err != nil {
		return err
	}

	options := types.ImagePullOptions{}
	log.Infof("pulling image %s", ref)
	res, err := cli.ImagePull(context.Background(), ref, options)

	// If you do not read from the response, ImagePull do nothing
	buf := new(bytes.Buffer)
	buf.ReadFrom(res)
	res.Close()

	return err
}

func ContainerCreate(name string) (types.ContainerCreateResponse, error) {
	cli, err := newClient()
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	config := &container.Config{
		Cmd:   []string{"/bin/bash"},
		Image: name,
	}
	hostConfig := &container.HostConfig{}
	networkingConfig := &network.NetworkingConfig{}

	log.Infof("creating container %s", name)
	res, err := cli.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, name)
	return res, err
}

func ContainerStart(id string) error {
	cli, err := newClient()
	if err != nil {
		return err
	}

	options := types.ContainerStartOptions{}
	log.Infof("starting container %s", id)
	return cli.ContainerStart(context.Background(), id, options)
}
