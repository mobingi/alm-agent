package api

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/mobingi/alm-agent/server_config"
)

func TestWriteTempToken(t *testing.T) {
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
	re := regexp.MustCompile(`ASIAX`)

	buf, err := ioutil.ReadFile(filepath.Join(awsConfDir, "credentials"))
	if err != nil {
		t.Fatal("credentials file was not created")
	}

	t.Log(string(buf))
	if !re.MatchString(string(buf)) {
		t.Fatal("credentials does not contain Key")
	}

	buf, err = ioutil.ReadFile(filepath.Join(awsConfDir, "awslogs_creds.conf"))
	if err != nil {
		t.Fatal("awslogs_creds file was not created")
	}

	t.Log(string(buf))
	if !re.MatchString(string(buf)) {
		t.Fatal("credentials does not contain Key")
	}
}

func TestFetServerConfigFromFile(t *testing.T) {
	sc := &serverConfig.Config{}
	getServerConfigFromFile("../test/fixtures/serverconfig.v1.json", sc)
	t.Log(sc)

	expected := "mobingi/ubuntu-apache2-php7:7.1"
	actual := sc.Image
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}
}
