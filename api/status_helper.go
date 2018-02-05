package api

import (
	"io/ioutil"

	"github.com/mobingi/alm-agent/util"
)

func isNewStatus(path, status string) bool {
	if !util.FileExists(path) {
		return true
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return true
	}

	if string(b) == status {
		return false
	}

	return true
}

func saveStatus(path, status string) error {
	err := ioutil.WriteFile(path, []byte(status), 0600)
	if err != nil {
		return err
	}
	return nil
}
