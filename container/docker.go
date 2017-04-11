package container

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingilabs/go-modaemon/server_config"
	"github.com/mobingilabs/go-modaemon/util"

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
	envs     []string
}

func NewDocker(s *serverConfig.Config) (*Docker, error) {
	docker := &Docker{
		image:    strings.TrimPrefix(s.Image, "http://"),
		username: s.DockerHubUserName,
		password: s.DockerHubPassword,
		ports:    s.Ports,
		pm:       portmapper.New(""),
		codeDir:  s.CodeDir,
		envs: func() []string {
			var envs []string
			for k, v := range s.EnvironmentVariables {
				es := []string{k, v}
				envs = append(envs, strings.Join(es, "="))
			}
			return envs
		}(),
	}

	chain := &iptables.ChainInfo{Name: "DOCKER", Table: "nat"}
	docker.pm.SetIptablesChain(chain, "docker0")

	defaultHeaders := map[string]string{"User-Agent": "modaemon"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.24", nil, defaultHeaders)
	docker.client = cli

	return docker, err
}

func (d *Docker) CheckImageUpdated() (bool, error) {
	res, err := d.imagePull()
	if err != nil {
		return false, err
	}

	if strings.Contains(res, "Image is up to date for") {
		return false, nil
	} else {
		return true, nil
	}
}

func (d *Docker) GetContainer(name string) (*Container, error) {
	filter := opts.NewFilterOpt()
	filter.Set(fmt.Sprintf("name=%s", name))
	options := types.ContainerListOptions{
		Filters: filter.Value(),
	}
	res, err := d.client.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, nil
	}

	name = strings.TrimPrefix(res[0].Names[0], "/")

	c := &Container{ID: res[0].ID, Name: name}
	c.IP, err = d.getIPAddress(c)
	if err != nil {
		return nil, err
	}
	return c, err
}

func (d *Docker) GetContainerIDbyImage(ancestor string) (string, error) {
	filter := opts.NewFilterOpt()
	filter.Set(fmt.Sprintf("ancestor=%s", ancestor))
	options := types.ContainerListOptions{
		Filters: filter.Value(),
	}
	res, err := d.client.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}

	return res[0].ID, nil
}

func (d *Docker) StartContainer(name string, dir string) (*Container, error) {

	_, err := d.imagePull()
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

func (d *Docker) UnmapPort() error {
	for _, port := range d.ports {
		key := &net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: port}
		err := d.pm.Unmap(key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Docker) RenameContainer(c *Container, name string) error {
	err := d.client.ContainerRename(context.Background(), c.ID, name)
	if err != nil {
		return err
	}
	c.Name = name
	return nil
}

func (d *Docker) getIPAddress(c *Container) (net.IP, error) {
	inspect, err := d.inspectContainer(c)
	if err != nil {
		return nil, err
	}
	return net.ParseIP(inspect.NetworkSettings.IPAddress), nil
}

func (d *Docker) imagePull() (string, error) {
	authConfig := &types.AuthConfig{
		Username: d.username,
		Password: d.password,
	}

	b, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	encodedAuth := base64.URLEncoding.EncodeToString(b)

	options := types.ImagePullOptions{
		RegistryAuth: encodedAuth,
	}
	log.Infof("pulling image %s", d.image)
	res, err := d.client.ImagePull(context.Background(), d.image, options)

	if err != nil {
		return "", err
	}

	// If you do not read from the response, ImagePull do nothing
	buf := new(bytes.Buffer)
	buf.ReadFrom(res)
	res.Close()

	return buf.String(), nil
}

func (d *Docker) containerCreate(name string, dir string) (*Container, error) {
	config := &container.Config{
		Image: d.image,
		Env:   d.envs,
	}
	log.Debugf("ContainerConfig: %#v", config)

	hostConfig := &container.HostConfig{}
	if dir != "" {
		bind := fmt.Sprintf("%s:%s", dir, d.codeDir)
		hostConfig.Binds = append(hostConfig.Binds, bind)

		initScriptFile := ""
		if util.FileExists(path.Join(dir, "mobingi-init.sh")) {
			initScriptFile = path.Join(dir, "mobingi-init.sh")
		} else if util.FileExists(path.Join(dir, "mobingi-install.sh")) {
			initScriptFile = path.Join(dir, "mobingi-install.sh")
		}

		if initScriptFile != "" {
			if !util.FileExists("/tmp/init") {
				if err := os.Mkdir("/tmp/init", 0700); err != nil {
					return nil, err
				}
			}

			if util.FileExists("/tmp/init/init.sh") {
				if err := os.Remove("/tmp/init/init.sh"); err != nil {
					return nil, err
				}
			}

			if err := os.Link(initScriptFile, "/tmp/init/init.sh"); err != nil {
				return nil, err
			}

			if err := os.Chmod("/tmp/init/init.sh", 0755); err != nil {
				return nil, err
			}

			hostConfig.Binds = append(hostConfig.Binds, "/tmp/init:/tmp/init")
		}
	}

	networkingConfig := &network.NetworkingConfig{}

	log.Infof("creating container \"%s\" from image \"%s\"", name, d.image)
	res, err := d.client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, name)
	log.Debugf("hostConfig: %#v", hostConfig)
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

func (d *Docker) CreateContainerExec(id string, cmd []string) *types.ContainerExecCreateResponse, error {
	return d.client.ContainerExecCreate(context.Background(), id, cmd)
}

func (d *Docker) StartContainerExec(id string, esc types.ExecStartCheck) *types.ContainerExecCreateResponse, error {
	return d.client.ContainerExecStart(context.Background(), id, esc)
}
