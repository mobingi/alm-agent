package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/server_config"
)

type conf struct {
	ServerConfig *serverConfig.Config
}

type client struct {
	client    *http.Client
	config    *config.Config
	tokenType string
	token     string
}

func NewClient(conf *config.Config) (*client, error) {
	c := &client{
		config: conf,
		client: &http.Client{},
	}

	return c, nil
}

func (c *client) getServerConfig() (*serverConfig.Config, error) {
	values := url.Values{}
	values.Set("stack_id", c.config.StackID)

	res, err := c.get("/v2/alm/serverconfig", values)
	if err != nil {
		return nil, err
	}

	conf := &conf{}
	err = json.Unmarshal(res, conf)
	if err != nil {
		return nil, err
	}

	return conf.ServerConfig, nil
}

func (c *client) get(path string, values url.Values) ([]byte, error) {
	req, err := http.NewRequest("GET", c.config.APIHost+path, nil)
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", c.tokenType, c.token))

	req.URL.RawQuery = values.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)
	return res, nil
}

func (c *client) getAccessToken() error {
	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", c.config.StackID)
	values.Set("client_secret", c.config.AuthorizationToken)

	resp, err := c.client.PostForm(c.config.APIHost+"/v2/access_token", values)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var res map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}

	c.tokenType = res["token_type"].(string)
	c.token = res["access_token"].(string)

	return nil
}
