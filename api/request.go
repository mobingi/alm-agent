package api

import (
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingi/alm-agent/server_config"
)

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
		err = getServerConfigFromFile(u.Path, conf)
		if err != nil {
			return nil, err
		}
	default:
		err = getServerConfigFromAPI(conf)
		if err != nil {
			return nil, err
		}
	}

	return conf, nil
}

func getServerConfigFromAPI(sc *serverConfig.Config) error {
	log.Debug("Step: serverConfig.getFromAPI")

	values := url.Values{}
	values.Set("stack_id", c.getConfig().StackID)

	log.Debug("Step: api: /v2/alm/serverconfig")
	err := Get("/v2/alm/serverconfig", values, sc)
	if err != nil {
		return err
	}

	return nil
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
	if err != nil {
		return err
	}
	return nil
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
