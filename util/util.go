package util

import (
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
)

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
