package api

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestWriteTempToken(t *testing.T) {
	tmpAWSDir, _ := ioutil.TempDir("", "containerLogs")
	awsConfDir = filepath.Join(tmpAWSDir, ".aws")
	defer os.RemoveAll(tmpAWSDir)

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
