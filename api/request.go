package api

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	"github.com/mobingi/alm-agent/metavars"
	"github.com/mobingi/alm-agent/server_config"
	"github.com/mobingi/alm-agent/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type route struct {
	AccessToken,
	EventSpotShutdown,
	InstanceStatus,
	AgentStatus,
	ContainerStatus,
	ServerConfig,
	Sts,
	LogsSTS string
}

// RoutesV2 points API v2. to figure out what has been implemented.
var RoutesV2 = &route{
	AccessToken:       "/v2/access_token",
	EventSpotShutdown: "/v2/event/spot/shutdown",
	InstanceStatus:    "/v2/alm/instance/status",
	ServerConfig:      "/v2/alm/serverconfig",
	Sts:               "/v2/alm/sts",
}

// RoutesV3 points API v3. to figure out what has been implemented.
var RoutesV3 = &route{
	AccessToken:       "/v3/access_token",
	EventSpotShutdown: "/v3/event/spot/shutdown",
	AgentStatus:       "/v3/alm/agent/agent_status",
	ContainerStatus:   "/v3/alm/agent/container_status",
	ServerConfig:      "/v3/alm/agent/config",
	Sts:               "/v3/alm/sts",
	LogsSTS:           "/v3/alm/agent/logs_access_token",
}

var tokenCachePath = "/opt/mobingi/etc/tokencache.json"
var lastContainerStatusPath = "/opt/mobingi/etc/last_container_status"
var lastAgentStatusPath = "/opt/mobingi/etc/last_agent_status"

// GetAccessToken requests token of user for auth by API.
func GetAccessToken() error {
	err := fetchAccessTokenCache()
	if err == nil {
		return nil
	}
	log.Debugf("GetAccessToken CacheState: %#v", err)

	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", c.getConfig().StackID)
	values.Set("client_secret", c.getConfig().AuthorizationToken)

	err = Post(RoutesV3.AccessToken, values, &apitoken)
	if err != nil {
		flushAccessTokenCache()
		return err
	}
	createAccessTokenCache()
	return nil
}

func createAccessTokenCache() {
	apitoken.ExpiresAt = time.Now().Unix() + apitoken.ExpiresIn
	at, _ := json.Marshal(apitoken)
	ioutil.WriteFile(tokenCachePath, []byte(at), 0600)
	return
}

func flushAccessTokenCache() {
	if util.FileExists(tokenCachePath) {
		os.Remove(tokenCachePath)
	}
	return
}

func fetchAccessTokenCache() error {
	at, err := ioutil.ReadFile(tokenCachePath)
	if err != nil {
		return err
	}
	json.Unmarshal([]byte(at), &apitoken)
	log.Debugf("fetchAccessTokenCache: %#v", apitoken)

	// reuse in 7hours
	if apitoken.ExpiresAt-time.Now().Unix() < 25200 {
		flushAccessTokenCache()
		return errors.New("apitoken should be renew")
	}

	log.Debug("use local cache of AccessToken")
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
	values.Set("flag", c.getConfig().Flag)

	err := Get(RoutesV3.ServerConfig, values, sc)
	if err != nil {
		return err
	}

	return nil
}

// GetStsToken to STS token for CWLogs
func GetStsToken() error {
	err := stsToken.fetchCache()
	if err == nil {
		return nil
	}
	log.Debug(err)

	values := url.Values{}
	values.Set("stack_id", c.getConfig().StackID)
	values.Set("service", "logs")

	err = Get(RoutesV3.Sts, values, &stsToken)
	if err != nil {
		return err
	}

	stsToken.createCache()
	err = stsToken.writeTempToken()
	if err != nil {
		stsToken.flushCache()
		return err
	}
	return nil
}

// SendAgentStatus send agent status to API
func SendAgentStatus(status, message string) error {
	if !isNewStatus(lastAgentStatusPath, status) {
		log.Debug("SendAgentStatus: skipped. agent status did not changed.")
		return nil
	}
	values := url.Values{}
	values.Set("stack_id", c.getConfig().StackID)
	values.Set("agent_id", metavars.AgentID)
	values.Set("status", status)
	if message != "" {
		values.Set("message", message)
	}

	if metavars.ServerID != "" {
		values.Set("instance_id", metavars.ServerID)
	}

	err := Post(RoutesV3.AgentStatus, values, nil)
	if err != nil {
		return err
	}
	saveStatus(lastAgentStatusPath, status)
	return nil
}

// SendContainerStatus send container app status to API
func SendContainerStatus(status string) error {
	if !isNewStatus(lastContainerStatusPath, status) {
		log.Debug("SendContainerStatus: skipped. container status did not changed.")
		return nil
	}

	values := url.Values{}
	values.Set("stack_id", c.getConfig().StackID)
	values.Set("agent_id", metavars.AgentID)
	values.Set("container_id", metavars.ServerID)
	values.Set("status", status)

	if metavars.ServerID != "" {
		values.Set("instance_id", metavars.ServerID)
	}

	err := Post(RoutesV3.ContainerStatus, values, nil)
	if err != nil {
		return err
	}
	saveStatus(lastContainerStatusPath, status)
	return nil
}

// SendInstanceStatus send container app status to API
// only V2
func SendInstanceStatus(status string) error {
	// use for debug enviromnent
	if metavars.ServerID == "" {
		log.Warnf("Skiped sending status to API(serverid is empty): %s", status)
		return nil
	}

	values := url.Values{}
	values.Set("instance_id", metavars.ServerID)
	values.Set("stack_id", c.getConfig().StackID)
	values.Set("status", status)

	err := Post(RoutesV2.InstanceStatus, values, nil)
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

	err := Post(RoutesV3.EventSpotShutdown, values, nil)
	return err
}
