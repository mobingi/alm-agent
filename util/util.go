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

func GetServerID(s ...string) (string, error) {
	var sid string

	if len(s) == 0 {
		s = append(s, "aws")
	}

	switch s[0] {
	case "aws":
		sid = getServerIDforEC2()
	default:
		return sid, errors.New("Provider `" + s[0] + "` is not supported.")
	}

	return sid, nil
}

func getServerIDforEC2() string {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(ec2METAENDPOINT + "/latest/meta-data/instance-id")
	if err != nil {
		log.Warnf("%#v", err)
		return ""
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("%#v", err)
		return ""
	}

	return string(body)
}
