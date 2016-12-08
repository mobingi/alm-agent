package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/mobingilabs/go-modaemon/config"
)

type Client struct {
	client    *http.Client
	config    *config.Config
	tokenType string
	token     string
}

func NewClient(conf *config.Config) (*Client, error) {
	c := &Client{
		config: conf,
		client: &http.Client{},
	}

	return c, nil
}

func (c *Client) getAccessToken() error {
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
