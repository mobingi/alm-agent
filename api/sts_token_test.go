package api

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mobingi/alm-agent/util"
	"github.com/stretchr/testify/assert"
)

func testCreateCache(t *testing.T) {
	assert := assert.New(t)
	stsToken := &StsToken{
		AccessKeyID:     "testKEY",
		SecretAccessKey: "testSKEY",
		SessionToken:    "testTOKEN",
	}
	tmpCacheDir, _ := ioutil.TempDir("", "stsToken")
	defer os.RemoveAll(tmpCacheDir)

	origstsTokenCachePath := stsTokenCachePath
	stsTokenCachePath = tmpCacheDir
	defer func() { stsTokenCachePath = origstsTokenCachePath }()
	stsToken.createCache()

	assert.True(util.FileExists(stsTokenCachePath))
}
