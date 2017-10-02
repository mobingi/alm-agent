package container

import (
	"net"
	"strings"

	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/server_config"
	log "github.com/sirupsen/logrus"

	client "docker.io/go-docker"
	"github.com/docker/libnetwork/iptables"
	"github.com/docker/libnetwork/portmapper"
)

// Docker is manager of docker
type Docker struct {
	Client   *client.Client
	Image    string
	Username string
	Password string
	Ports    []int
	Pm       *portmapper.PortMapper
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
		Pm:       portmapper.New(""),
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

	chain := &iptables.ChainInfo{Name: "DOCKER", Table: "nat"}
	docker.Pm.SetIptablesChain(chain, "docker0")

	defaultHeaders := map[string]string{"User-Agent": defaultUA}
	cli, err := client.NewClient(dockerSock, dockerAPIVer, nil, defaultHeaders)
	docker.Client = cli

	return docker, err
}

// MapPort allocates listner
func (d *Docker) MapPort(c *Container) error {
	for _, port := range d.Ports {
		dest := &net.TCPAddr{IP: c.IP, Port: port}
		_, err := d.Pm.Map(dest, net.IPv4(0, 0, 0, 0), port, true)
		if err != nil {
			return err
		}
		log.Infof("MapPort: %d", port)
	}
	return nil
}

// UnmapPort disallocates listner
func (d *Docker) UnmapPort() error {
	for _, port := range d.Ports {
		key := &net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: port}
		err := d.Pm.Unmap(key)
		if err != nil {
			return err
		}
		log.Infof("UnmapPort: %d", port)
	}
	return nil
}
