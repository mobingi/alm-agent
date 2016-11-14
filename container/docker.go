package container

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingilabs/go-modaemon/server_config"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/opts"
	"github.com/docker/libnetwork/iptables"
	"github.com/docker/libnetwork/portmapper"
	"golang.org/x/net/context"
)

type Docker struct {
	client   *client.Client
	image    string
	username string
	password string
	ports    []int
	pm       *portmapper.PortMapper
	codeDir  string
}

func NewDocker(s serverConfig.Config) (*Docker, error) {
	docker := &Docker{
		image:    strings.TrimPrefix(s.Image, "http://"),
		username: s.DockerHubUserName,
		password: s.DockerHubPassword,
		ports:    s.Ports,
		pm:       portmapper.New(""),
		codeDir:  s.CodeDir,
	}

	chain := &iptables.ChainInfo{Name: "DOCKER", Table: "nat"}
	docker.pm.SetIptablesChain(chain, "docker0")

	defaultHeaders := map[string]string{"User-Agent": "modaemon"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.24", nil, defaultHeaders)
	docker.client = cli

	return docker, err
}

func (d *Docker) GetContainer(name string) (*Container, error) {
	filter := opts.NewFilterOpt()
	filter.Set(fmt.Sprintf("name=%s", name))
	options := types.ContainerListOptions{
		Filter: filter.Value(),
	}
	res, err := d.client.ContainerList(context.Background(), options)

	name = strings.TrimPrefix(res[0].Names[0], "/")

	c := &Container{ID: res[0].ID, Name: name}
	c.IP, err = d.getIPAddress(c)
	if err != nil {
		return nil, err
	}
	return c, err
}

func (d *Docker) StartContainer(name string, dir string) (*Container, error) {

	err := d.imagePull()
	if err != nil {
		return nil, err
	}

	c, err := d.containerCreate(name, dir)
	if err != nil {
		return nil, err
	}

	err = d.containerStart(c)
	if err != nil {
		return nil, err
	}

	c.IP, err = d.getIPAddress(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (d *Docker) MapPort(c *Container) error {
	for _, port := range d.ports {
		dest := &net.TCPAddr{IP: c.IP, Port: port}
		_, err := d.pm.Map(dest, net.IPv4(0, 0, 0, 0), port, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Docker) UnmapPort(c *Container) error {
	for _, port := range d.ports {
		key := &net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: port}
		err := d.pm.Unmap(key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Docker) getIPAddress(c *Container) (net.IP, error) {
	inspect, err := d.inspectContainer(c)
	if err != nil {
		return nil, err
	}
	return net.ParseIP(inspect.NetworkSettings.IPAddress), nil
}

func (d *Docker) imagePull() error {
	authConfig := &types.AuthConfig{
		Username: d.username,
		Password: d.password,
	}

	b, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}
	encodedAuth := base64.URLEncoding.EncodeToString(b)

	options := types.ImagePullOptions{
		RegistryAuth: encodedAuth,
	}
	log.Infof("pulling image %s", d.image)
	res, err := d.client.ImagePull(context.Background(), d.image, options)

	if err != nil {
		return err
	}

	// If you do not read from the response, ImagePull do nothing
	buf := new(bytes.Buffer)
	buf.ReadFrom(res)
	res.Close()

	return nil
}

func (d *Docker) containerCreate(name string, dir string) (*Container, error) {
	config := &container.Config{
		Image: d.image,
	}

	hostConfig := &container.HostConfig{}
	bind := fmt.Sprintf("%s:%s", dir, d.codeDir)
	hostConfig.Binds = append(hostConfig.Binds, bind)

	networkingConfig := &network.NetworkingConfig{}

	log.Infof("creating container \"%s\" from image \"%s\"", name, d.image)
	res, err := d.client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, name)
	return &Container{Name: name, ID: res.ID}, err
}

func (d *Docker) containerStart(c *Container) error {
	options := types.ContainerStartOptions{}
	log.Infof("starting container %s", c.ID)
	return d.client.ContainerStart(context.Background(), c.ID, options)
}

func (d *Docker) inspectContainer(c *Container) (types.ContainerJSON, error) {
	return d.client.ContainerInspect(context.Background(), c.ID)
}

func (d *Docker) StopContainer(c *Container) error {
	timeout := 3 * time.Second
	return d.client.ContainerStop(context.Background(), c.ID, &timeout)
}

func (d *Docker) RemoveContainer(c *Container) error {
	options := types.ContainerRemoveOptions{}
	return d.client.ContainerRemove(context.Background(), c.ID, options)
}
