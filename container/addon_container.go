package container

import (
	"github.com/mobingi/alm-agent/server_config"

	client "docker.io/go-docker"
	"github.com/mobingi/alm-agent/config"
)

// NewAddonDocker returns docker client
func NewAddonDocker(c *config.Config, name string, opts serverConfig.Addon, ac *SystemContainer) (*Docker, error) {
	envs := []string{}
	for _, envfunc := range ac.EnvFuncs {
		envs = append(envs, handleEnvMap[envfunc](c, opts)...)
	}

	docker := &Docker{
		Image: ac.Image,
		Envs:  envs,
	}

	defaultHeaders := map[string]string{"User-Agent": defaultUA}
	cli, err := client.NewClient(dockerSock, dockerAPIVer, nil, defaultHeaders)
	docker.Client = cli

	return docker, err
}
