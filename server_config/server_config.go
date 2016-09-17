package serverConfig

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Config struct {
	DockerImageName string
}

func Get(url url.URL) (Config, error) {
	resp, err := http.Get(url.String())
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
