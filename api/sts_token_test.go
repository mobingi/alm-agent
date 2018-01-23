package api

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mobingi/alm-agent/util"
	"github.com/stretchr/testify/assert"
)

func TestCreateCache(t *testing.T) {
	assert := assert.New(t)
	stsToken := &StsToken{
		AccessKeyID:     "testKEY",
		SecretAccessKey: "testSKEY",
		SessionToken:    "testTOKEN",
	}
	tmpCacheDir, _ := ioutil.TempDir("", "stsToken")
	defer os.RemoveAll(tmpCacheDir)

	origstsTokenCachePath := stsTokenCachePath
	stsTokenCachePath = filepath.Join(tmpCacheDir, "sts_cache.json")
	defer func() { stsTokenCachePath = origstsTokenCachePath }()
	stsToken.createCache()

	assert.True(util.FileExists(stsTokenCachePath))
}

func TestFetchCache(t *testing.T) {
	assert := assert.New(t)
	stsToken := &StsToken{
		AccessKeyID:     "testKEY",
		SecretAccessKey: "testSKEY",
		SessionToken:    "testTOKEN",
	}
	tmpCacheDir, _ := ioutil.TempDir("", "stsToken")
	defer os.RemoveAll(tmpCacheDir)

	origstsTokenCachePath := stsTokenCachePath
	stsTokenCachePath = filepath.Join(tmpCacheDir, "sts_cache.json")
	defer func() { stsTokenCachePath = origstsTokenCachePath }()

	err := stsToken.fetchCache()
	assert.EqualError(err, "no STS cache")

	stsToken.createCache()

	err = stsToken.fetchCache()
	assert.Nil(err)

	assert.Equal("testKEY", stsToken.AccessKeyID)
	assert.Equal("testSKEY", stsToken.SecretAccessKey)
	assert.Equal("testTOKEN", stsToken.SessionToken)

	// refresh
	now := time.Now()
	atime := now.Add(-(time.Duration(30) * time.Minute))
	mtime := now.Add(-(time.Duration(30) * time.Minute))
	os.Chtimes(stsTokenCachePath, atime, mtime)

	err = stsToken.fetchCache()
	assert.EqualError(err, "STS cache too old, will be renew")
}

func TestWriteTempToken(t *testing.T) {
	assert := assert.New(t)
	tmpAWSDir, _ := ioutil.TempDir("", "TestWriteTempToken")
	defer os.RemoveAll(tmpAWSDir)

	origawsConfDir := awsConfDir
	awsConfDir = filepath.Join(tmpAWSDir, ".aws")
	defer func() { awsConfDir = origawsConfDir }()

	sts := &StsToken{
		AccessKeyID:     "ASIAXXXXXXXXXXXXXXX",
		SecretAccessKey: "SAXXX",
		SessionToken:    "STSTOKENXXX",
	}
	sts.writeTempToken()

	buf, err := ioutil.ReadFile(filepath.Join(awsConfDir, "awslogs_creds.conf"))
	if err != nil {
		t.Fatal("awslogs_creds file was not created")
	}

	assert.Contains(string(buf), "ASIAX")
	assert.Contains(string(buf), "STSTOKENXXX")
}
