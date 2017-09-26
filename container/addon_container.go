package container

import (
	"github.com/mobingi/alm-agent/server_config"

	"github.com/docker/docker/client"
	"github.com/mobingi/alm-agent/config"
)

// AddonContainer uses log
type AddonContainer struct {
	Name     string
	Image    string
	EnvFuncs []string
	VolFuncs []string
}

// NewAddonDocker returns docker client
func NewAddonDocker(c *config.Config, opts *serverConfig.Addon, ac *AddonContainer) (*Docker, error) {
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
