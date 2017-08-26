package container

import (
	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/metavars"
)

// SystemContainer uses log
type SystemContainer struct {
	Name     string
	Image    string
	EnvFuncs []string
	VolFuncs []string
}

// SystemContainers is slice of SystemContainer
type SystemContainers struct {
	Container []SystemContainer
}

// EnvHandle builds string to append to ENV.
type EnvHandle func(*config.Config) string

// EnvFuncs contains funcs of EnvHandle
type EnvFuncs struct{}

// VolHandle builds []string to volume mount.
type VolHandle func() []string

// VolFuncs contains funcs of VolHandle
type VolFuncs struct{}

var envFuncs = &EnvFuncs{}
var volFuncs = &VolFuncs{}

var handleEnvMap = map[string]EnvHandle{
	"stack_id":    envFuncs.StackID,
	"instance_id": envFuncs.InstanceID,
}

var handleVolMap = map[string]VolHandle{
	"logs_vol": volFuncs.LogsVol,
}

// NewSysDocker returns docker client
func NewSysDocker(c *config.Config, s *SystemContainer) (*Docker, error) {
	envs := []string{}
	for _, envfunc := range s.EnvFuncs {
		envs = append(envs, handleEnvMap[envfunc](c))
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

// StackID resolves stack_id
func (e *EnvFuncs) StackID(c *config.Config) string {
	return "STACK_ID=" + c.StackID
}

// InstanceID resolves instance_id
func (e *EnvFuncs) InstanceID(c *config.Config) string {
	return "INSTANCE_ID=" + metavars.ServerID
}

// LogsVol resolves logs_vol
func (v *VolFuncs) LogsVol() []string {
	vols := []string{
		"/root/.aws/awslogs_creds.conf:/etc/awslogs/awscli.conf",
		"/var/log:/var/log",
		containerLogsLocation + ":/var/container",
		"/opt/awslogs:/var/lib/awslogs",
	}
	return vols
}

// StartSysContainer starts docker container
func (d *Docker) StartSysContainer(name string, s *SystemContainer) (*Container, error) {

	_, err := d.imagePull()
	if err != nil {
		return nil, err
	}

	c, err := d.sysContainerCreate(name, s)
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

func (d *Docker) sysContainerCreate(name string, s *SystemContainer) (*Container, error) {
	config := &container.Config{
		Image: d.Image,
		Env:   d.Envs,
	}
	log.Debugf("ContainerConfig: %#v", config)

	hostConfig := &container.HostConfig{}

	vols := []string{}
	for _, envfunc := range s.VolFuncs {
		vols = append(vols, handleVolMap[envfunc]()...)
	}

	hostConfig.Binds = append(
		hostConfig.Binds,
		vols...,
	)

	networkingConfig := &network.NetworkingConfig{}
	log.Infof("creating container \"%s\" from image \"%s\"", name, d.Image)
	res, err := d.Client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, name)
	log.Debugf("hostConfig: %#v", hostConfig)

	return &Container{Name: name, ID: res.ID}, err
}
