package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/mobingi/alm-agent/util"
	log "github.com/sirupsen/logrus"
)

var stsTokenCachePath = "/opt/mobingi/etc/sts_cache.json"

// StsToken for grant access for AWS resources
type StsToken struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// use Cache 40-55min randomly
func (sts *StsToken) fetchCache() error {
	if util.FileExists(stsTokenCachePath) {
		rand.Seed(time.Now().UnixNano())
		t := time.Now()
		it := 55 - rand.Intn(15)
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
