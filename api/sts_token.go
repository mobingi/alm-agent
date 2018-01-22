package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/mobingi/alm-agent/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var awsConfDir = "/root/.aws"
var stsTokenCachePath = "/opt/mobingi/etc/sts_cache.json"

// StsToken for grant access for AWS resources
type StsToken struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// use Cache 20-29min randomly
func (sts *StsToken) fetchCache() error {
	if util.FileExists(stsTokenCachePath) {
		rand.Seed(time.Now().UnixNano())
		t := time.Now()
		it := 29 - rand.Intn(9)
		at := t.Add(-(time.Duration(it) * time.Minute))

		file, _ := os.Open(stsTokenCachePath)
		defer file.Close()

		info, _ := file.Stat()
		if at.After(info.ModTime()) {
			return errors.New("STS cache too old, will be renew")
		}
		log.Debug("use local cache of StsToken")
		cache, _ := ioutil.ReadFile(stsTokenCachePath)
		json.Unmarshal([]byte(cache), &sts)
		return nil
	}
	return errors.New("no STS cache")
}

func (sts *StsToken) createCache() error {
	token, _ := json.Marshal(sts)
	ioutil.WriteFile(stsTokenCachePath, []byte(token), 0600)
	return nil
}

func (sts *StsToken) flushCache() {
	if util.FileExists(stsTokenCachePath) {
		os.Remove(stsTokenCachePath)
	}
	return
}

func (sts *StsToken) writeTempToken() error {
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

	logscreadsContent := fmt.Sprintf(creadsForlogs, sts.AccessKeyID, sts.SecretAccessKey, sts.SessionToken, region)

	err := ioutil.WriteFile(filepath.Join(awsConfDir, "awslogs_creds.conf"), []byte(logscreadsContent), 0600)
	if err != nil {
		return errors.Wrap(err, "failed to write awslogs_creds.conf.")
	}
	return nil
}
