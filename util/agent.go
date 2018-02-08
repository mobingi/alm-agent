package util

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/docker/distribution/uuid"
	"github.com/mobingi/alm-agent/metavars"
	log "github.com/sirupsen/logrus"
)

var (
	METAENDPOINT          = "http://169.254.169.254/"
	ecsMETAENDPOINT       = "http://100.100.100.200/"
	gceMETAENDPOINT       = "http://metadata.google.internal/"
	containerLogsLocation = "/var/log/alm-agent/container"
	agentIDSavePath       = "/opt/mobingi/etc/alm-agent.id"
)

type k5 struct {
	Uuid string `json:"uuid"`
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
		endpoint = METAENDPOINT + "/latest/meta-data/instance-id"
	case "alicloud":
		endpoint = ecsMETAENDPOINT + "/latest/meta-data/instance-id"
	case "gcp":
		endpoint = gceMETAENDPOINT + "/computeMetadata/v1/instance/id"
	case "k5":
		endpoint = METAENDPOINT + "/openstack/latest/meta_data.json"
	case "localtest":
		return "", nil
	default:
		return "", errors.New("Provider `" + provider + "` is not supported.")
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Warnf("%#v", err)
		return "", errors.New("Failed to return Request")
	}

	if provider == "gcp" {
		req.Header.Set("Metadata-Flavor", "Google")
	}

	resp, err := client.Do(req)
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

	if provider == "k5" {
		id, err := getUuidOfK5(body)
		if err != nil {
			return "", errors.New("Failed to get ServerID")
		}
		return id, nil
	}

	return string(body), nil
}

func getUuidOfK5(b []byte) (string, error) {
	var k k5
	if err := json.Unmarshal(b, &k); err != nil {
		return "", err
	}
	return k.Uuid, nil
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
