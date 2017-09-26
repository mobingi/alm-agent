package container

import (
	"github.com/mobingi/alm-agent/server_config"

	"github.com/docker/docker/client"
	"github.com/mobingi/alm-agent/config"
)

// NewAddonDocker returns docker client
func NewAddonDocker(c *config.Config, opts *serverConfig.Addon, s *SystemContainer) (*Docker, error) {
	envs := []string{}
	for _, envfunc := range s.EnvFuncs {
		envs = append(envs, handleEnvMap[envfunc](c, opts)...)
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
