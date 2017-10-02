package container

import (
	"strings"

	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/server_config"

	client "docker.io/go-docker"
)

// Docker is manager of docker
type Docker struct {
	Client   *client.Client
	Image    string
	Username string
	Password string
	Ports    []int
	Pm       interface{}
	CodeDir  string
	Envs     []string
}

// NewDocker is construcor for DockerClient
func NewDocker(c *config.Config, s *serverConfig.Config) (*Docker, error) {
	docker := &Docker{
		Image:    strings.TrimPrefix(s.Image, "http://"),
		Username: s.DockerHubUserName,
		Password: s.DockerHubPassword,
		Ports:    s.Ports,
		CodeDir:  s.CodeDir,
		Envs: func() []string {
			var Envs []string
			Envs = append(Envs, "MO_USER_ID="+c.UserID, "MO_STACK_ID="+c.StackID)
			for k, v := range s.EnvironmentVariables {
				es := []string{k, v}
				Envs = append(Envs, strings.Join(es, "="))
			}
			return Envs
		}(),
	}

	defaultHeaders := map[string]string{"User-Agent": defaultUA}
	cli, err := client.NewClient(dockerSock, dockerAPIVer, nil, defaultHeaders)
	docker.Client = cli

	return docker, err
}

// MapPort allocates listner
func (d *Docker) MapPort(c *Container) error {
	return nil
}

// UnmapPort disallocates listner
func (d *Docker) UnmapPort() error {
	return nil
}
