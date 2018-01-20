package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/mobingi/alm-agent/server_config"
	"github.com/mobingi/alm-agent/util"
	log "github.com/sirupsen/logrus"
)

var awsConfDir = "/root/.aws"

// writeTempToken o save STS token for CWLogs container
func writeTempToken() error {
	region := logregion

	creadsForlogs := `[plugins]
cwlogs = cwlogs
[default]
aws_access_key_id=%s
aws_secret_access_key=%s
aws_session_token=%s
region=%s
`

	if !util.FileExists(awsConfDir) {
		os.Mkdir(awsConfDir, 0700)
	}

	logscreadsContent := fmt.Sprintf(creadsForlogs, ststoken.AccessKeyID, ststoken.SecretAccessKey, ststoken.SessionToken, region)

	err := ioutil.WriteFile(filepath.Join(awsConfDir, "awslogs_creds.conf"), []byte(logscreadsContent), 0600)
	if err != nil {
		return errors.Wrap(err, "failed to write awslogs_creds.conf.")
	}
	return nil
}

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
