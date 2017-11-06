package serverConfig

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/mobingi/alm-agent/util"
	log "github.com/sirupsen/logrus"
)

// PubKey for SSH
type PubKey struct {
	PublicKey string
}

// Addon list of enabled addons
type Addon interface{}

// Config means application stack
type Config struct {
	Image                string
	DockerHubUserName    string
	DockerHubPassword    string
	CodeDir              string
	GitReference         string
	GitRepo              string
	GitPrivateKey        string
	Ports                []int
	Updated              uint
	Users                map[string]*PubKey
	EnvironmentVariables map[string]string
	Addons               []Addon `json:"addon"`
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

	log.Debugf("updated of %s is %s", versionPath, fu)
	log.Debugf("updated of serverconfig is %s", c.Updated)

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
