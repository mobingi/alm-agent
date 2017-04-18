package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/server_config"
	"github.com/mobingilabs/go-modaemon/util"
)

type client struct {
	client    *http.Client
	config    *config.Config
	tokenType string
	token     string
}

type StsToken struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func NewClient(conf *config.Config) (*client, error) {
	c := &client{
		config: conf,
		client: &http.Client{},
	}

	err := c.getAccessToken()

	return c, err
}

func (c *client) GetServerConfig(sclocation string) (*serverConfig.Config, error) {
	u, err := url.Parse(sclocation)
	if err != nil {
		return nil, err
	}
	log.Debugf("%#v", u)

	conf := &serverConfig.Config{}
	switch u.Scheme {
	case "file":
		log.Debug("Step: serverConfig.getFromFile")
		b, err := ioutil.ReadFile(u.Path)
		if err != nil {
			return nil, err
		}

		log.Debugf("SCFfromfile: %s", b)

		err = json.Unmarshal(b, conf)
		if err != nil {
			return nil, err
		}
	default:
		log.Debug("Step: serverConfig.getFromHTTP")

		values := url.Values{}
		values.Set("stack_id", c.config.StackID)

		log.Debug("Step: api: /v2/alm/serverconfig")
		res, err := c.get("/v2/alm/serverconfig", values)
		if err != nil {
			return nil, err
		}
		log.Debugf("Response: %s", res)
		err = json.Unmarshal(res, conf)
		if err != nil {
			return nil, err
		}
	}

	return conf, nil
}

func (c *client) GetStsToken() (*StsToken, error) {
	values := url.Values{}
	values.Set("user_id", c.config.UserID)
	values.Set("stack_id", c.config.StackID)

	res, err := c.get("/v2/alm/sts", values)
	if err != nil {
		return nil, err
	}

	stsToken := &StsToken{}
	err = json.Unmarshal(res, stsToken)
	if err != nil {
		return nil, err
	}

	return stsToken, nil
}

func (c *client) WriteTempToken(token *StsToken) error {
	creadsTemplate := `[tempcreds]
aws_access_key_id=%s
aws_secret_access_key=%s
aws_session_token=%s
`
	if !util.FileExists("/root/.aws") {
		os.Mkdir("/root/.aws", 0700)
	}

	creadsContent := fmt.Sprintf(creadsTemplate, token.AccessKeyID, token.SecretAccessKey, token.SessionToken)
	err := ioutil.WriteFile("/root/.aws/credentials", []byte(creadsContent), 0600)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) SendInstanceStatus(serverID, status string) error {
	values := url.Values{}
	values.Set("instance_id", serverID)
	values.Set("stack_id", c.config.StackID)
	values.Set("status", status)

	_, err := c.post("/v2/alm/instance/status", values)
	return err
}

func (c *client) SendSpotShutdownEvent(serverID string) error {
	values := url.Values{}
	values.Set("user_id", c.config.UserID)
	values.Set("stack_id", c.config.StackID)
	values.Set("instance_id", serverID)

	_, err := c.post("/v2/event/spot/shutdown", values)
	return err
}

func (c *client) get(path string, values url.Values) ([]byte, error) {
	req, err := http.NewRequest("GET", c.config.APIHost+path, nil)
	if c.token != "" && c.tokenType != "" {
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", c.tokenType, c.token))
	}

	req.URL.RawQuery = values.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return res, errors.New(resp.Status)
	} else {
		return res, nil
	}
}

func (c *client) post(path string, values url.Values) ([]byte, error) {
	req, err := http.NewRequest("POST", c.config.APIHost+path, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.token != "" && c.tokenType != "" {
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", c.tokenType, c.token))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return res, errors.New(resp.Status)
	} else {
		return res, nil
	}
}

func (c *client) getAccessToken() error {
	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", c.config.StackID)
	values.Set("client_secret", c.config.AuthorizationToken)

	res, err := c.post("/v2/access_token", values)
	if err != nil {
		return err
	}

	var tokenInfo map[string]interface{}

	err = json.Unmarshal(res, &tokenInfo)
	if err != nil {
		return err
	}

	c.tokenType = tokenInfo["token_type"].(string)
	c.token = tokenInfo["access_token"].(string)

	return nil
}
