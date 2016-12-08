package config

import (
	"encoding/json"
	"io/ioutil"
)

// Config is a struct which has config data from modaemon.cfg
type Config struct {
	UserID             string
	StackID            string
	LogicalStakID      string
	AccessKey          string
	SecretKey          string
	APIHost            string
	AuthorizationToken string
	StorageService     string
	LogBucket          string
	ServerRole         string
}

func Load(b []byte) (Config, error) {
	c := Config{}
	err := json.Unmarshal(b, &c)
	return c, err
}

func LoadFromFile(file string) (Config, error) {
	c := Config{}

	dat, err := ioutil.ReadFile(file)

	if err != nil {
		return c, err
	}

	return Load(dat)
}
