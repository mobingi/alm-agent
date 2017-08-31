package config

import (
	"encoding/json"
	"io/ioutil"
)

// Config is a struct which has config data from alm-agent.cfg
type Config struct {
	UserID             string
	StackID            string
	APIHost            string
	AuthorizationToken string
	Flag               string
}

// Load just do Unmarshal
func Load(b []byte) (*Config, error) {
	c := &Config{}
	err := json.Unmarshal(b, &c)

	return c, err
}

// LoadFromFile just do read
func LoadFromFile(file string) (*Config, error) {
	c := &Config{}

	dat, err := ioutil.ReadFile(file)

	if err != nil {
		return c, err
	}

	return Load(dat)
}
