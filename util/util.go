package util

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
)

var METAENDPOINT = "http://169.254.169.254/"

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func FetchContainerState() string {
	var container_status string = "/var/modaemon/containerlogs/log/container_status"
	if !FileExists(container_status) {
		return ""
	}

	dat, err := ioutil.ReadFile(container_status)
	if err != nil {
		return ""
	}

	log.Debugf("FetchContainerState: %s", string(dat))
	return string(dat)
}

func GetServerID(s ...string) (string, error) {
	var sid string

	if len(s) == 0 {
		s[0] = "aws"
	}

	switch s[0] {
	case "aws":
		sid = getServerIDforEC2()
	default:
		return sid, errors.New("Provider is not supported.")
	}

	return sid, nil
}

func getServerIDforEC2() string {
	resp, err := http.Get(METAENDPOINT + "/latest/meta-data/instance-id")
	if err != nil {
		log.Fatalf("%#v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("%#v", err)
	}
	return string(body)
}
