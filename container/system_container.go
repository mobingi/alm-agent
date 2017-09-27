package container

import (
	"context"
	"os"

	"github.com/mobingi/alm-agent/metavars"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/mobingi/alm-agent/config"
	log "github.com/sirupsen/logrus"
)

// SystemContainer uses log
type SystemContainer struct {
	Name     string
	Image    string
	EnvFuncs []string
	VolFuncs []string
	Restart  bool
}

// SystemContainers is slice of SystemContainer
type SystemContainers struct {
	Container []SystemContainer
}

// NewSysDocker returns docker client
func NewSysDocker(c *config.Config, s *SystemContainer) (*Docker, error) {
	envs := []string{}
	for _, envfunc := range s.EnvFuncs {
		envs = append(envs, handleEnvMap[envfunc](c, nil)...)
	}

	docker := &Docker{
		Image: s.Image,
		Envs:  envs,
	}

	defaultHeaders := map[string]string{"User-Agent": defaultUA}
	cli, err := client.NewClient(dockerSock, dockerAPIVer, nil, defaultHeaders)
	docker.Client = cli

	return docker, err
}

// StartSysContainer starts docker container
func (d *Docker) StartSysContainer(s *SystemContainer) (*Container, error) {

	_, err := d.imagePull()
	if err != nil {
		return nil, err
	}

	c, err := d.sysContainerCreate(s)
	if err != nil {
		return nil, err
	}

	err = d.containerStart(c)
	if err != nil {
		return nil, err
	}

	ct, _ := d.Client.ContainerInspect(context.Background(), c.ID)
	if err == nil {
		log.Debugf("ContainerInspect: %#v", ct)
	}

	cp, _ := d.Client.ContainerStatPath(context.Background(), c.ID, "/")
	if err == nil {
		log.Debugf("ContainerInspect: %#v", cp)
	}

	c.IP, err = d.getIPAddress(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (d *Docker) sysContainerCreate(s *SystemContainer) (*Container, error) {
	config := &container.Config{
		Image: d.Image,
		Env:   d.Envs,
	}

	if metavars.ServerID == "" {
		config.Hostname, _ = os.Hostname()
	} else {
		config.Hostname = metavars.ServerID
	}

	log.Debugf("ContainerConfig: %#v", config)

	hostConfig := &container.HostConfig{}

	vols := []string{}
	for _, volfunc := range s.VolFuncs {
		vols = append(vols, handleVolMap[volfunc]()...)
	}

	hostConfig.Binds = append(
		hostConfig.Binds,
		vols...,
	)

	networkingConfig := &network.NetworkingConfig{}
	log.Infof("creating container \"%s\" from image \"%s\"", s.Name, d.Image)
	res, err := d.Client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, s.Name)
	log.Debugf("hostConfig: %#v", hostConfig)

	return &Container{Name: s.Name, ID: res.ID}, err
}
