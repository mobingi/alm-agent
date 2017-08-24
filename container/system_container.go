package container

import (
	"github.com/derekparker/delve/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// NewSysDocker returns docker client
func NewSysDocker(c *config.Config, id string) (*container.Docker, error) {
	docker := &container.Docker{
		Image: "mobingi/mo-awslogs",
		Envs:  []string{"STACK_ID=" + c.StackID, "INSTANCE_ID=" + id},
	}

	defaultHeaders := map[string]string{"User-Agent": defaultUA}
	cli, err := client.NewClient(dockerSock, dockerAPIVer, nil, defaultHeaders)
	docker.Client = cli

	return docker, err
}
