package serverConfig

import (
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingi/alm-agent/util"
)

// PubKey for SSH
type PubKey struct {
	PublicKey string
}

// Config means application stack
type Config struct {
	Image                string
	DockerHubUserName    string
	DockerHubPassword    string
	Code                 string
	CodeDir              string
	GitReference         string
	GitPrivateKey        string
	Ports                []int
	Updated              uint
	Users                map[string]*PubKey
	EnvironmentVariables map[string]string
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

	fu := strings.Trim(string(dat), "\n")
	cu := fmt.Sprintf("%d", c.Updated)

	log.Debugf("updated of %s is %s", versionPath, fu)
	log.Debugf("updated of serverconfig is %s", cu)

	if fu < cu {
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
