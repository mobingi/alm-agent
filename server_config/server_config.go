package serverConfig

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Config struct {
	Image             string
	DockerHubUserName string
	DockerHubPassword string
	Code              string
	CodeDir           string
	GitReference      string
	Ports             []int
}

func Get(url string) (Config, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Config{}, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Config{}, err
	}

	return parse(b)
}

func parse(b []byte) (Config, error) {
	c := Config{}
	err := json.Unmarshal(b, &c)
	return c, err
}
