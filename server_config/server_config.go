package serverConfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingilabs/go-modaemon/util"
)

type PubKey struct {
	PublicKey string
}

type Config struct {
	Image                string
	DockerHubUserName    string
	DockerHubPassword    string
	Code                 string
	CodeDir              string
	GitReference         string
	GitPrivateKey        string
	Ports                []int
	Updated              int
	Users                map[string]*PubKey
	EnvironmentVariables map[string]string
}

var path = "/var/modaemon/updated"

func NeedsUpdate(c *Config) (bool, error) {
	if !util.FileExists(path) {
		return true, nil
	}

	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}

	fu := strings.Trim(string(dat), "\n")
	cu := fmt.Sprintf("%d", c.Updated)

	log.Debugf("updated of %s is %s", path, fu)
	log.Debugf("updated of serverconfig is %s", cu)

	if fu != cu {
		return true, nil
	}

	log.Debug("No need to do task")
	return false, nil
}

func WriteUpdated(c *Config) error {
	if util.FileExists(path) {
		if err := os.Remove(path); err != nil {
			return err
		}
	}

	v := fmt.Sprintf("%d", c.Updated)
	err := ioutil.WriteFile(path, []byte(v), 0644)
	if err != nil {
		return err
	}

	return nil
}
