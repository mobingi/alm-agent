package config

import (
	"encoding/json"
	"io/ioutil"
)

const configFile = "/opt/modaemon/modaemon.cfg"

// Config is a struct which has config data from modaemon.cfg
type Config struct {
	ServerConfigAPIEndPoint string
	UserID                  string
	StackID                 string
	LogicalStakID           string
	AccessKey               string
	SecretKey               string
	APIHost                 string
	AuthorizationToken      string
	StorageService          string
	LogBucket               string
	ServerRole              string
	HideAWSFromContainers   string
}

func Load(b []byte) (Config, error) {
	c := Config{}
	err := json.Unmarshal(b, &c)
	return c, err
}

func LoadFromFile() (Config, error) {
	c := Config{}

	dat, err := ioutil.ReadFile(configFile)

	if err != nil {
		return c, err
	}

	return Load(dat)
}
