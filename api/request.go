package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingi/alm-agent/server_config"
	"github.com/mobingi/alm-agent/util"
)

// GetServerConfig retrives serverconfig from file or API.
func GetServerConfig(sclocation string) (*serverConfig.Config, error) {
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
		values.Set("stack_id", c.getConfig().StackID)

		log.Debug("Step: api: /v2/alm/serverconfig")
		res := Get("/v2/alm/serverconfig", values, &conf)
		log.Debugf("Response: %s", res)
	}

	return conf, nil
}

// GetStsToken to STS token for CWLogs
func GetStsToken() (*StsToken, error) {
	values := url.Values{}
	values.Set("user_id", c.getConfig().UserID)
	values.Set("stack_id", c.getConfig().StackID)

	err := Get("/v2/alm/sts", values, &stsToken)
	if err != nil {
		return nil, err
	}

	return &stsToken, nil
}

// WriteTempToken to save STS token for CWLogs container
func WriteTempToken(token *StsToken) error {
	region := logregion

	creadsTemplate := `[tempcreds]
aws_access_key_id=%s
aws_secret_access_key=%s
aws_session_token=%s
region=%s
`

	creadsForlogs := `[plugins]
cwlogs = cwlogs
[default]
aws_access_key_id=%s
aws_secret_access_key=%s
aws_session_token=%s
region=%s
`

	if !util.FileExists("/root/.aws") {
		os.Mkdir("/root/.aws", 0700)
	}

	tempcreadsContent := fmt.Sprintf(creadsTemplate, token.AccessKeyID, token.SecretAccessKey, token.SessionToken, region)
	logscreadsContent := fmt.Sprintf(creadsForlogs, token.AccessKeyID, token.SecretAccessKey, token.SessionToken, region)
	err := ioutil.WriteFile("/root/.aws/credentials", []byte(tempcreadsContent), 0600)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("/root/.aws/awslogs_creds.conf", []byte(logscreadsContent), 0600)
	if err != nil {
		return err
	}
	return nil
}

// SendInstanceStatus send container app status to API
func SendInstanceStatus(serverID, status string) error {
	values := url.Values{}
	values.Set("instance_id", serverID)
	values.Set("stack_id", c.getConfig().StackID)
	values.Set("status", status)

	// use for debug enviromnent
	if serverID == "" {
		log.Warnf("Skiped sending status to API(serverid is empty): %s", status)
		return nil
	}

	err := Post("/v2/alm/instance/status", values, nil)
	return err
}

// SendSpotShutdownEvent nortifies that instance detects shutdown event.
func SendSpotShutdownEvent(serverID string) error {
	values := url.Values{}
	values.Set("user_id", c.getConfig().UserID)
	values.Set("stack_id", c.getConfig().StackID)
	values.Set("instance_id", serverID)

	err := Post("/v2/event/spot/shutdown", values, nil)
	return err
}

// GetAccessToken requests token of user for auth by API.
func GetAccessToken() error {
	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", c.getConfig().StackID)
	values.Set("client_secret", c.getConfig().AuthorizationToken)

	err := Post("/v2/access_token", values, &apitoken)
	if err != nil {
		return err
	}
	return nil
}
