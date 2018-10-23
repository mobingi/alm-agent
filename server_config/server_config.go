package serverConfig

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/mobingi/alm-agent/shared_volume"
	"github.com/mobingi/alm-agent/util"
	log "github.com/sirupsen/logrus"
)

// ContainerUser for SSH
type ContainerUser struct {
	UserName  string `json:"name"`
	PublicKey string `json:"publicKey"`
}

// Addon list of enabled addons
type Addon interface{}

// Config means application stack
type Config struct {
	Image                string                     `json:"container_image"`
	DockerHubUserName    string                     `json:"container_registry_username"`
	DockerHubPassword    string                     `json:"container_registry_password"`
	CodeDir              string                     `json:"container_code_dir"`
	GitReference         string                     `json:"container_git_reference"`
	GitRepo              string                     `json:"container_git_repo"`
	GitPrivateKey        string                     `json:"container_git_private_key"`
	Ports                []int                      `json:"container_ports"`
	Updated              uint                       `json:"container_updated"`
	Users                []ContainerUser            `json:"container_users"`
	EnvironmentVariables map[string]string          `json:"container_env_vars"`
	Addons               []Addon                    `json:"container_addons"`
	SharedVolume         *sharedvolume.SharedVolume `json:"container_storage"`
}

var versionPath = "/opt/mobingi/etc/configVersion"

// NeedsUpdate checks latest serverconfig
func NeedsUpdate(c *Config) (bool, error) {
	if !util.FileExists(versionPath) {
		return true, nil
	}

	dat, err := ioutil.ReadFile(versionPath)
	if err != nil {
		return false, err
	}

	fu, _ := strconv.ParseUint(string(dat), 10, 32)

	log.Debugf("updated of %s is %d", versionPath, fu)
	log.Debugf("updated of serverconfig is %d", c.Updated)

	if uint(fu) < c.Updated {
		return true, nil
	}

	log.Debug("No need to do task")
	return false, nil
}

// WriteUpdated stores current
func WriteUpdated(c *Config) error {
	v := fmt.Sprintf("%d", c.Updated)
	log.Debugf("Write %s to %s", v, versionPath)
	err := ioutil.WriteFile(versionPath, []byte(v), 0644)
	if err != nil {
		return err
	}

	return nil
}
