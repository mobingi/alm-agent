package api

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mobingi/alm-agent/server_config"
	"github.com/stretchr/testify/assert"
)

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
	WriteTempToken(sts)

	buf, err := ioutil.ReadFile(filepath.Join(awsConfDir, "awslogs_creds.conf"))
	if err != nil {
		t.Fatal("awslogs_creds file was not created")
	}

	assert.Contains(string(buf), "ASIAX")
	assert.Contains(string(buf), "STSTOKENXXX")
}

func TestGetServerConfigFromFile(t *testing.T) {
	assert := assert.New(t)
	sc := &serverConfig.Config{}
	getServerConfigFromFile("../test/fixtures/serverconfig.v2.json", sc)

	assert.Equal(sc.Image, "mobingi/ubuntu-apache2-php7:7.1")
	assert.NotEmpty(sc.EnvironmentVariables)
}
