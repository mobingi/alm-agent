package util

import (
	"io/ioutil"
	"strings"

	"github.com/docker/distribution/uuid"
	"github.com/mobingi/alm-agent/metavars"
	log "github.com/sirupsen/logrus"
)

var (
	awsMETAENDPOINT       = "http://169.254.169.254/"
	k5METAENDPOINT        = awsMETAENDPOINT
	ecsMETAENDPOINT       = "http://100.100.100.200/"
	gceMETAENDPOINT       = "http://metadata.google.internal/"
	containerLogsLocation = "/var/log/alm-agent/container"
	agentIDSavePath       = "/opt/mobingi/etc/alm-agent.id"
)

type k5 struct {
	UUID string `json:"uuid"`
}

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
func GetServerID(provider string) error {
	p, err := newProvider(provider)
	if err != nil {
		return err
	}

	sid, err := p.GetServerID()
	if err != nil {
		return err
	}

	metavars.ServerID = sid
	return nil
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
