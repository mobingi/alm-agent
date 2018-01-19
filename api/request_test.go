package api

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"testing"

	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/server_config"
	log "github.com/sirupsen/logrus"
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
	flushTokenCache(tokenCachePath)
	defer flushTokenCache(tokenCachePath)
	var apitoken = &apiToken{}
	tc := &config.Config{
		APIHost:            "https://test.example.com",
		StackID:            "teststack",
		AuthorizationToken: "testtoken",
	}

	SetConfig(tc)
	origPost := Post
	defer func() { Post = origPost }()
	mockPost(
		func(path string, values url.Values, target interface{}) error {
			res := `{
				"token_type": "Bearer",
				"access_token": "eyJ0eXAiOiJKV1",
				"expires_in": 43200
			}`
			err := json.Unmarshal([]byte(res), &apitoken)
			if err != nil {
				t.Fatal("Failed Unmarshal into apiToken.")
			}
			return nil
		},
	)

	err := GetAccessToken()
	if err != nil {
		t.Fatal("Failed GetAccessToken.")
	}

	expected := "Bearer"
	actual := apitoken.TokenType
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}

	expectedInt := int64(43200)
	actualInt := apitoken.ExpiresIn
	if actualInt != expectedInt {
		t.Fatalf("Expected: %d\n But: %d", expectedInt, actualInt)
	}
}

func TestGetSTSToken(t *testing.T) {
	flushTokenCache(tokenCachePath)
	defer flushTokenCache(tokenCachePath)
	log.SetLevel(log.DebugLevel)
	var testtoken = &StsToken{}
	tc := &config.Config{
		APIHost:            "https://test.example.com",
		StackID:            "teststack",
		AuthorizationToken: "testtoken",
	}

	SetConfig(tc)
	origPost := Post
	defer func() { Post = origPost }()
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

func TestGetServerConfigFromAPI(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	tc := &config.Config{
		APIHost:            "https://test.example.com",
		StackID:            "teststack",
		AuthorizationToken: "testtoken",
	}

	SetConfig(tc)

	var sc = &serverConfig.Config{}
	origGet := Get
	defer func() { Get = origGet }()
	mockGet(
		func(path string, values url.Values, target interface{}) error {
			res, _ := ioutil.ReadFile("../test/fixtures/serverconfig.v2.json")
			err := json.Unmarshal([]byte(res), &sc)
			if err != nil {
				t.Log(err)
				t.Fatal("Failed Unmarshal into ServerConfig.")
			}
			return nil
		},
	)

	err := getServerConfigFromAPI(sc)
	if err != nil {
		t.Fatal("Failed getServerConfigFromAPI.")
	}
	t.Log(sc)

	expected := "mobingi/ubuntu-apache2-php7:7.1"
	actual := sc.Image
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}
}
