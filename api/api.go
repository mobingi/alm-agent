package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingi/alm-agent/config"
)

var (
	logregion = "ap-northeast-1"
	c         clientInterface
	apitoken  apiToken
	stsToken  StsToken
)

type clientInterface interface {
	buildURI(string) string
	getHTTPClient() *http.Client
	setConfig(*config.Config) error
	getConfig() *config.Config
}

type apiToken struct {
	TokenType string `json:"token_type"`
	Token     string `json:"access_token"`
}

// StsToken for CWLogs
type StsToken struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

type client struct {
	config *config.Config
}

func (c *client) getHTTPClient() *http.Client {
	return &http.Client{}
}

func (c *client) buildURI(path string) string {
	return c.getConfig().APIHost + path
}

// SetConfig updates client.config.
func SetConfig(conf *config.Config) error {
	c.setConfig(conf)
	return nil
}

func (c *client) setConfig(conf *config.Config) error {
	c.config = conf
	return nil
}

func (c *client) getConfig() *config.Config {
	return c.config
}

func init() {
	log.Debug("Initializing api client...")
	c = &client{}
}

// Get wraps HTTP GET Request
var Get = func(path string, values url.Values, target interface{}) error {
	log.Debugf("Get: %s", path)
	log.Debugf("%#v", c.getConfig())
	log.Debugf("%#v", apitoken)
	req, err := http.NewRequest("GET", c.buildURI(path), nil)

	if apitoken.Token != "" && apitoken.TokenType != "" {
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", apitoken.TokenType, apitoken.Token))
	}
	log.Debugf("%#v", req)

	req.URL.RawQuery = values.Encode()
	httpClient := c.getHTTPClient()
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	log.Debugf("%#v", string(res))
	err = json.Unmarshal(res, &target)
	if err != nil {
		return err
	}

	return nil
}

// Post wraps HTTP Post Request
var Post = func(path string, values url.Values, target interface{}) error {
	log.Debugf("Post: %s", path)
	log.Debugf("%#v", c.getConfig())
	req, err := http.NewRequest("POST", c.buildURI(path), strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if apitoken.Token != "" && apitoken.TokenType != "" {
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", apitoken.TokenType, apitoken.Token))
	}
	log.Debugf("%#v", req)

	httpClient := c.getHTTPClient()
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	log.Debugf("%#v", string(res))
	err = json.Unmarshal(res, &target)
	if err != nil {
		return err
	}

	return nil
}
