package util

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	ec2METAENDPOINT       = "http://169.254.169.254/"
	ecsMETAENDPOINT       = "http://100.100.100.200/"
	containerLogsLocation = "/var/modaemon/containerlogs"
)

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

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

func GetServerID(s string) (string, error) {
	sid, err := getServerID(s)
	if err != nil {
		return "", err
	}

	return sid, nil
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
	default:
		return "", errors.New("Provider `" + provider + "` is not supported.")
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		log.Warnf("%#v", err)
		return "", errors.New("Faild to get ServerID")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("%#v", err)
		return "", errors.New("Faild to get ServerID")
	}

	return string(body), nil
}
