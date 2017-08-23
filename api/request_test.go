package api

import (
	"encoding/json"
	"net/url"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingi/alm-agent/config"
)

func mockGet(fn func(path string, values url.Values, target interface{}) error) {
	Get = fn
	return
}

func mockPost(fn func(path string, values url.Values, target interface{}) error) {
	Post = fn
	return
}

func TestGetAccessToken(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	var testtoken = &StsToken{}
	tc := &config.Config{
		APIHost:            "https://test.example.com",
		StackID:            "teststack",
		AuthorizationToken: "testtoken",
	}

	SetConfig(tc)
	mockPost(
		func(path string, values url.Values, target interface{}) error {
			res := `{
				"AccessKeyId": "ASIAXXXXXXXXXXXXXXX",
				"SecretAccessKey": "SAXXX",
				"SessionToken": "STSTOKENXXX"
			}`
			err := json.Unmarshal([]byte(res), &testtoken)
			if err != nil {
				t.Fatal("Failed Unmarshal into StsToken.")
			}
			return nil
		},
	)

	err := GetAccessToken()
	if err != nil {
		t.Fatal("Failed GetAccessToken.")
	}

	expected := "ASIAXXXXXXXXXXXXXXX"
	actual := testtoken.AccessKeyID
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}

	expected = "SAXXX"
	actual = testtoken.SecretAccessKey
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}

	expected = "STSTOKENXXX"
	actual = testtoken.SessionToken
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}
}
