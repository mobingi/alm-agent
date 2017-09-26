package util

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/docker/distribution/uuid"
	"github.com/mobingi/alm-agent/metavars"
)

var (
	ec2METAENDPOINT       = "http://169.254.169.254/"
	ecsMETAENDPOINT       = "http://100.100.100.200/"
	containerLogsLocation = "/var/log/alm-agent/container"
	agentIDSavePath       = "/opt/mobingi/etc/alm-agent.id"
)

// FetchContainerState fetches state of application in running container.
func FetchContainerState() string {
	containerStatus := containerLogsLocation + "/log/container_status"
	if !FileExists(containerStatus) {
		return ""
	}

	dat, err := ioutil.ReadFile(containerStatus)
	if err != nil {
		return ""
	}

	log.Debugf("FetchContainerState: %s", string(dat))
	return strings.TrimSpace(string(dat))
}

// GetServerID returns string that identify VM on running provider. (e.g. instance ID)
func GetServerID(s string) error {
	sid, err := getServerID(s)
	if err != nil {
		return err
	}

	metavars.ServerID = sid
	return nil
}

func getServerID(provider string) (string, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	var endpoint string
	switch provider {
	case "aws":
		endpoint = ec2METAENDPOINT + "/latest/meta-data/instance-id"
	case "alicloud":
		endpoint = ecsMETAENDPOINT + "/latest/meta-data/instance-id"
	case "localtest":
		return "", nil
	default:
		return "", errors.New("Provider `" + provider + "` is not supported.")
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		log.Warnf("%#v", err)
		return "", errors.New("Failed to get ServerID")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("%#v", err)
		return "", errors.New("Failed to get ServerID")
	}

	return string(body), nil
}

// AgentID sets metavars.AgentID
func AgentID() {
	if FileExists(agentIDSavePath) {
		dat, _ := ioutil.ReadFile(agentIDSavePath)
		metavars.AgentID = string(dat)
		return
	}
	id := uuid.Generate()
	ioutil.WriteFile(agentIDSavePath, []byte(id.String()), 0644)
	metavars.AgentID = id.String()
	return
}
