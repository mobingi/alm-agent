package container

import (
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

type SystemContainers struct {
	Container []SystemContainer
}

type EnvHandle func(*config.Config) string
type EnvFuncs struct{}

var envFuncs = &EnvFuncs{}

var handleEnvMap = map[string]EnvHandle{
	"stack_id":    envFuncs.StackID,
	"instance_id": envFuncs.InstanceID,
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
