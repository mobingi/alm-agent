package container

import (
	"github.com/docker/docker/client"
	"github.com/mobingi/alm-agent/config"
)

// NewSysDocker returns docker client
func NewSysDocker(c *config.Config, id string) (*Docker, error) {
	docker := &Docker{
		Image: "mobingi/alm-awslogs",
		Envs:  []string{"STACK_ID=" + c.StackID, "INSTANCE_ID=" + id},
	}

	defaultHeaders := map[string]string{"User-Agent": defaultUA}
	cli, err := client.NewClient(dockerSock, dockerAPIVer, nil, defaultHeaders)
	docker.Client = cli

	return docker, err
}
