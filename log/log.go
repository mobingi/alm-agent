package log

import (
	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/container"

	dcontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

func NewDocker(c *config.Config, id string) (*container.Docker, error) {
	docker := &Docker{
		image: "mobingi/mo-awslogs",
		envs: func() []string {
			envs := []string{"STACK_ID=" + c.StackID, "INSTANCE_ID=" + id}
			return envs
		}(),
	}

	defaultHeaders := map[string]string{"User-Agent", "mo-awslogs"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.24", nil, defaultHeaders)
	docker.client = cli

	return docker, err
}

func (d *container.Docker) StartContainer(name string) (*container.Container, error) {
	_, err := d.container.ImagePull()
	if err != nil {
		return nil, err
	}

	c, err := d.containerCreate(name)
	if err != nil {
		return nil, err
	}

	err = d.container.ContainerStart(c)
	if err != nil {
		return nil, err
	}

	ct, _ := d.client.ContainerInspect(context.Background(), c.ID)
	if err == nil {
		log.Debugf("ContainerInspect: %#v", ct)
	}

	cp, _ := d.client.ContainerStatPath(context.Background(), c.ID, "/")
	if err == nil {
		log.Debugf("ContainerInspect: %#v", cp)
	}

	return c, nil
}

func (d *container.Docker) containerCreate(name string) (*container.Container, error) {
	config := &dcontainer.Config{
		Image: d.image,
		Env:   d.envs,
	}
	log.Debugf("ContainerConfig: %#v", config)

	hostConfig := &dcontainer.HostConfig{}
	hostConfig.Binds = append(
		hostConfig.Binds,
		"/root/.aws/awslogs_creds.conf:/etc/awslogs/awscli.conf",
		"/var/log:/var/log",
		"/var/modaemon/containerlogs:/var/containerlogs",
		"/opt/awslogs:/var/lib/awslogs",
	)

	networkingConfig := &network.NetworkingConfig{}

	log.Infof("creating container \"%s\" from image \"%s\"", name, d.image)
	res, err := d.client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, name)
	log.Debugf("hostConfig: %#v", hostConfig)

	return &container.Container{Name: name, ID: res.ID}, err
}
