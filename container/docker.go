package container

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mobingi/alm-agent/shared_volume"
	"github.com/mobingi/alm-agent/util"
	log "github.com/sirupsen/logrus"

	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"docker.io/go-docker/api/types/filters"
	"docker.io/go-docker/api/types/network"
	"golang.org/x/net/context"
)

var (
	dockerAPIVer          = "v1.24"
	dockerSock            = "unix:///var/run/docker.sock"
	containerLogsLocation = "/var/log/alm-agent/container"
	defaultUA             = "mobingi alm-agent"
)

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
	args := filters.NewArgs(
		filters.KeyValuePair{
			Key:   "name",
			Value: name,
		},
	)
	//	args.Add("name", name)
	options := types.ContainerListOptions{
		Filters: args,
		All:     true,
	}
	res, err := d.Client.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, nil
	}

	name = strings.TrimPrefix(res[0].Names[0], "/")

	c := &Container{ID: res[0].ID, Name: name, State: res[0].State}
	c.IP, err = d.getIPAddress(c)
	if err != nil {
		return nil, err
	}
	return c, err
}

// GetContainerIDbyImage returns container by Image
func (d *Docker) GetContainerIDbyImage(ancestor string) (string, error) {
	args := filters.NewArgs(
		filters.KeyValuePair{
			Key:   "ancestor",
			Value: ancestor,
		},
	)
	options := types.ContainerListOptions{
		Filters: args,
		All:     true,
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

// prepareSharedVolume sets up volume
func (d *Docker) prepareSharedVolume(volumesetting *sharedvolume.SharedVolume) error {
	var v sharedvolume.Interface

	// empty volumesetting
	if volumesetting == nil {
		log.Debug("prepareSharedVolume: no settings")
		v = &sharedvolume.NullVolume{}
		d.SharedVolume = ""
		return nil
	}

	// return errors.New("Faild to setup shared volume")
	switch volumesetting.Type {
	case "efs":
		log.Debug("prepareSharedVolume: found efs setting")
		var mountpath = volumesetting.MountPath
		if volumesetting.MountPath == "" {
			// fallback to default
			mountpath = sharedvolume.DefarultMouhtPath
		}

		v = &sharedvolume.EFSVolume{
			Client: d.Client,
			Name:   "efsvolume",
			EFSID:  volumesetting.Identifier,
		}
		d.SharedVolume = fmt.Sprintf("efsvolume:%s", mountpath)
	case "local":
		log.Debug("prepareSharedVolume: found efs setting")
		var mountpath = volumesetting.MountPath
		if volumesetting.MountPath == "" {
			// fallback to default
			mountpath = sharedvolume.DefarultMouhtPath
		}

		v = &sharedvolume.LocalVolume{
			Client: d.Client,
			Name:   "localvolume",
		}
		d.SharedVolume = fmt.Sprintf("localvolume:%s", mountpath)
	default:
		log.Debug("prepareSharedVolume: no settings")
		v = &sharedvolume.NullVolume{}
		d.SharedVolume = ""
		return nil
	}

	err := v.Setup()
	if err != nil {
		return err
	}

	return nil
}

// cannot use sharedvolume.EFSVolume literal (type sharedvolume.EFSVolume) as type *sharedvolume.Interface in assignment:
// 	*sharedvolume.Interface is pointer to interface, not interface

// StartContainer starts docker container
func (d *Docker) StartContainer(name string, dir string) (*Container, error) {
	log.Infof("pulling image %s", d.Image)
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
		return c, err
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
		return c, err
	}

	return c, nil
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
	log.Debugf("pulling image %s", d.Image)
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

func (d *Docker) containerCreate(name string, dir string) (*Container, error) {
	config := &container.Config{
		Image: d.Image,
		Env:   d.Envs,
	}
	log.Debugf("ContainerConfig: %#v", config)

	hostConfig := &container.HostConfig{}
	// 1gb
	hostConfig.ShmSize = 1073741824
	hostConfig.Sysctls = map[string]string{
		"net.core.somaxconn":           "40960",
		"net.ipv4.ip_local_port_range": "10240 65535",
	}

	if dir != "" {
		codedir := d.fetchCodeDir()
		bind := fmt.Sprintf("%s:%s", dir, codedir)
		hostConfig.Binds = append(hostConfig.Binds, bind)

		initScriptFile := ""
		if util.FileExists(filepath.Join(dir, "mobingi-init.sh")) {
			initScriptFile = filepath.Join(dir, "mobingi-init.sh")
		} else if util.FileExists(filepath.Join(dir, "mobingi-install.sh")) {
			initScriptFile = filepath.Join(dir, "mobingi-install.sh")
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

		if d.SharedVolume != "" {
			hostConfig.Binds = append(hostConfig.Binds, d.SharedVolume)
		}
	}

	bindLog := containerLogsLocation + "/log:/var/log"
	hostConfig.Binds = append(hostConfig.Binds, bindLog)

	networkingConfig := &network.NetworkingConfig{}

	log.Infof("creating container \"%s\" from image \"%s\"", name, d.Image)
	res, err := d.Client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, name)
	log.Debugf("hostConfig: %#v", hostConfig)
	return &Container{Name: name, ID: res.ID}, err
}

func (d *Docker) containerStart(c *Container) error {
	options := types.ContainerStartOptions{}
	log.Infof("starting container %s", c.ID)
	return d.Client.ContainerStart(context.Background(), c.ID, options)
}

func (d *Docker) inspectContainer(c *Container) (types.ContainerJSON, error) {
	return d.Client.ContainerInspect(context.Background(), c.ID)
}

// ContainerHealth returns nil or ContainerState.Health
func (d *Docker) ContainerHealth(c *Container) (*types.Health, error) {
	ci, err := d.inspectContainer(c)
	return ci.State.Health, err
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

func (d *Docker) fetchCodeDir() string {
	if d.CodeDir != "" {
		return d.CodeDir
	}

	codeDirFromImage := d.getValueByLabelFromImage("com.mobingi.code_dir")
	if codeDirFromImage != "" {
		return codeDirFromImage
	}

	return "/var/www/html"
}

func (d *Docker) getValueByLabelFromImage(label string) string {
	args := filters.NewArgs(
		filters.Arg("reference", d.Image),
	)

	options := types.ImageListOptions{
		Filters: args,
	}

	ctx := context.Background()
	images, err := d.Client.ImageList(ctx, options)
	if err != nil {
		return ""
	}

	return images[0].Labels[label]
}
