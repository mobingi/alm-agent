package api

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/mobingi/alm-agent/server_config"
	log "github.com/sirupsen/logrus"
)

func getServerConfigFromFile(path string, sc *serverConfig.Config) error {
	log.Debugf("Step: serverConfig.getFromFile %s", path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "failed to serverConfig.getFromFile.")
	}

	log.Debugf("SCFfromfile: %s", b)
	err = json.Unmarshal(b, sc)
	if err != nil {
		return err
	}
	return nil
}
