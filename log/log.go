package log

import (
	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/container"

	"github.com/docker/docker/client"
)

func NewDocker(c *config.Config, id string) (*container.Docker, error) {
	docker := &container.Docker{
		Image: "mobingi/mo-awslogs",
		Envs: func() []string {
			Envs := []string{"STACK_ID=" + c.StackID, "INSTANCE_ID=" + id}
			return Envs
		}(),
	}

	defaultHeaders := map[string]string{"User-Agent": "mo-awslogs"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.24", nil, defaultHeaders)
	docker.Client = cli

	return docker, err
}
