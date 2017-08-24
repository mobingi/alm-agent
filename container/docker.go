package container

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/server_config"
	"github.com/mobingi/alm-agent/util"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/opts"
	"github.com/docker/libnetwork/iptables"
	"github.com/docker/libnetwork/portmapper"
	"golang.org/x/net/context"
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

var (
	dockerAPIVer          = "v1.24"
	dockerSock            = "unix:///var/run/docker.sock"
	containerLogsLocation = "/var/log/alm-agent/container"
	defaultUA             = "mobingi alm-agent"
)

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

// CheckImageUpdated pulls latest image if exsist.
func (d *Docker) CheckImageUpdated() (bool, error) {
	res, err := d.imagePull()
	if err != nil {
		return false, err
	}

	if strings.Contains(res, "Image is up to date for") {
		return false, nil
	}

	return true, nil
}

// GetContainer returns container by name
func (d *Docker) GetContainer(name string) (*Container, error) {
	filter := opts.NewFilterOpt()
	filter.Set(fmt.Sprintf("name=%s", name))
	options := types.ContainerListOptions{
		Filters: filter.Value(),
	}
	res, err := d.Client.ContainerList(context.Background(), options)
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

// GetContainerIDbyImage returns container by Image
func (d *Docker) GetContainerIDbyImage(ancestor string) (string, error) {
	filter := opts.NewFilterOpt()
	filter.Set(fmt.Sprintf("ancestor=%s", ancestor))
	options := types.ContainerListOptions{
		Filters: filter.Value(),
	}
	res, err := d.Client.ContainerList(context.Background(), options)
	if err != nil {
		return "", err
	}

	if len(res) < 1 {
		return "", nil
	}

	return res[0].ID, nil
}

// StartContainer starts docker container
func (d *Docker) StartContainer(name string, dir string, isApp bool) (*Container, error) {

	_, err := d.imagePull()
	if err != nil {
		return nil, err
	}

	c, err := d.containerCreate(name, dir, isApp)
	if err != nil {
		return nil, err
	}

	err = d.containerStart(c)
	if err != nil {
		return nil, err
	}

	ct, _ := d.Client.ContainerInspect(context.Background(), c.ID)
	if err == nil {
		log.Debugf("ContainerInspect: %#v", ct)
	}

	cp, _ := d.Client.ContainerStatPath(context.Background(), c.ID, "/")
	if err == nil {
		log.Debugf("ContainerInspect: %#v", cp)
	}

	c.IP, err = d.getIPAddress(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// MapPort allocates listner
func (d *Docker) MapPort(c *Container) error {
	for _, port := range d.Ports {
		dest := &net.TCPAddr{IP: c.IP, Port: port}
		_, err := d.Pm.Map(dest, net.IPv4(0, 0, 0, 0), port, true)
		if err != nil {
			return err
		}
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
	}
	return nil
}

// RenameContainer renames after lounch
func (d *Docker) RenameContainer(c *Container, name string) error {
	err := d.Client.ContainerRename(context.Background(), c.ID, name)
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
		Username: d.Username,
		Password: d.Password,
	}

	b, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	encodedAuth := base64.URLEncoding.EncodeToString(b)

	options := types.ImagePullOptions{
		RegistryAuth: encodedAuth,
	}
	log.Infof("pulling image %s", d.Image)
	res, err := d.Client.ImagePull(context.Background(), d.Image, options)

	if err != nil {
		return "", err
	}

	// If you do not read from the response, ImagePull do nothing
	buf := new(bytes.Buffer)
	buf.ReadFrom(res)
	res.Close()

	return buf.String(), nil
}

func (d *Docker) containerCreate(name string, dir string, isApp bool) (*Container, error) {
	if isApp {
		d.prepareLogsDir()
	}

	config := &container.Config{
		Image: d.Image,
		Env:   d.Envs,
	}
	log.Debugf("ContainerConfig: %#v", config)

	hostConfig := &container.HostConfig{}
	if dir != "" {
		bind := fmt.Sprintf("%s:%s", dir, d.CodeDir)
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

	if isApp {
		bindLog := containerLogsLocation + "/log:/var/log"
		hostConfig.Binds = append(hostConfig.Binds, bindLog)
	} else {
		hostConfig.Binds = append(
			hostConfig.Binds,
			"/root/.aws/awslogs_creds.conf:/etc/awslogs/awscli.conf",
			"/var/log:/var/log",
			containerLogsLocation+":/var/container",
			"/opt/awslogs:/var/lib/awslogs",
		)
	}

	networkingConfig := &network.NetworkingConfig{}

	log.Infof("creating container \"%s\" from image \"%s\"", name, d.Image)
	res, err := d.Client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, name)
	log.Debugf("hostConfig: %#v", hostConfig)
	return &Container{Name: name, ID: res.ID}, err
}

// to keep compatibility with older modaemon
func (d *Docker) prepareLogsDir() error {
	if util.FileExists(containerLogsLocation + "log") {
		return nil
	}

	log.Debug("prepareLogsDir: Start")
	ep := []string{
		"/bin/sh",
	}
	cmd := []string{
		"-c",
		"while true ; do sleep 1 ; done",
	}
	config := &container.Config{
		Image:      d.Image,
		Entrypoint: ep,
		Cmd:        cmd,
	}

	hostConfig := &container.HostConfig{}
	res, err := d.Client.ContainerCreate(context.Background(), config, hostConfig, &network.NetworkingConfig{}, "preparelogs")
	if err != nil {
		log.Errorf("prepareLogsDir.ContainerCreate: %#v", err)
	}

	options := types.ContainerStartOptions{}
	err = d.Client.ContainerStart(context.Background(), res.ID, options)
	if err != nil {
		log.Errorf("prepareLogsDir.ContainerStart: %#v", err)
	}

	os.MkdirAll(containerLogsLocation, 0755)

	err = exec.Command("docker", "cp", res.ID+":/var/log", containerLogsLocation).Run()
	if err != nil {
		log.Errorf("prepareLogsDir.copyFromContainerLogsLocation: %#v", err)
	}
	//	tmpcID := strings.TrimSpace(string(out))
	err = d.Client.ContainerKill(context.Background(), "preparelogs", "KILL")
	if err != nil {
		log.Errorf("prepareLogsDir.ContainerKill: %#v", err)
	}
	err = d.Client.ContainerRemove(context.Background(), "preparelogs", types.ContainerRemoveOptions{})
	if err != nil {
		log.Errorf("prepareLogsDir.ContainerRemove: %#v", err)
	}
	return nil
}

func (d *Docker) containerStart(c *Container) error {
	options := types.ContainerStartOptions{}
	log.Infof("starting container %s", c.ID)
	return d.Client.ContainerStart(context.Background(), c.ID, options)
}

func (d *Docker) inspectContainer(c *Container) (types.ContainerJSON, error) {
	return d.Client.ContainerInspect(context.Background(), c.ID)
}

// StopContainer stops contaner
func (d *Docker) StopContainer(c *Container) error {
	timeout := 3 * time.Second
	return d.Client.ContainerStop(context.Background(), c.ID, &timeout)
}

// RemoveContainer Removes stopped contaner
func (d *Docker) RemoveContainer(c *Container) error {
	options := types.ContainerRemoveOptions{}
	return d.Client.ContainerRemove(context.Background(), c.ID, options)
}

// CreateContainerExec prepaces exec on running container
func (d *Docker) CreateContainerExec(id string, cmd []string) (types.IDResponse, error) {
	exc := types.ExecConfig{
		Cmd: cmd,
	}
	return d.Client.ContainerExecCreate(context.Background(), id, exc)
}

// StartContainerExec do exec on running container
func (d *Docker) StartContainerExec(id string, esc types.ExecStartCheck) error {
	return d.Client.ContainerExecStart(context.Background(), id, esc)
}
