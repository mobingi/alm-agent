package docker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type Docker struct {
	client        *client.Client
	image         string
	username      string
	password      string
	identityToken string
}

func New(image, username, password string) (*Docker, error) {
	docker := &Docker{
		image:    strings.TrimPrefix(image, "http://"),
		username: username,
		password: password,
	}
	defaultHeaders := map[string]string{"User-Agent": "modaemon"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.24", nil, defaultHeaders)
	docker.client = cli

	return docker, err
}

func (d *Docker) ImagePull() error {
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

	// If you do not read from the response, ImagePull do nothing
	buf := new(bytes.Buffer)
	buf.ReadFrom(res)
	res.Close()

	return err
}

func (d *Docker) ContainerCreate(name string) (types.ContainerCreateResponse, error) {
	config := &container.Config{
		Cmd:   []string{"/bin/bash"},
		Image: name,
	}
	hostConfig := &container.HostConfig{}
	networkingConfig := &network.NetworkingConfig{}

	log.Infof("creating container %s", name)
	res, err := d.client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, name)
	return res, err
}

func (d *Docker) ContainerStart(id string) error {
	options := types.ContainerStartOptions{}
	log.Infof("starting container %s", id)
	return d.client.ContainerStart(context.Background(), id, options)
}
