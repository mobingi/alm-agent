package util

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// Provider is common interface for CloudVM
type Provider interface {
	MetadataEndpoint() string
	GetServerID() (string, error)
}

type provider struct{}

func (p *provider) simpleHTTPGet(endpoint string) (string, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	req, _ := http.NewRequest("GET", endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Warnf("%#v", err)
		return "", errors.New("Failed to get ServerID")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("%#v", err)
		return "", errors.New("Failed to get ServerID")
	}

	return string(body), nil
}

func (p *provider) simpleHTTPGetWithHeader(endpoint string, headers map[string]string) (string, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	req, _ := http.NewRequest("GET", endpoint, nil)
	for hKey, hVal := range headers {
		req.Header.Set(hKey, hVal)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Warnf("%#v", err)
		return "", errors.New("Failed to get ServerID")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("%#v", err)
		return "", errors.New("Failed to get ServerID")
	}

	return string(body), nil
}

func newProvider(name string) (Provider, error) {
	switch name {
	case "aws":
		return &awsProvider{}, nil
	case "alicloud":
		return &alicloudProvider{}, nil
	case "azure":
		return &azureProvider{}, nil
	case "gcp":
		return &gcpProvider{}, nil
	case "k5":
		return &k5Provider{}, nil
	case "localtest":
		return &nullProvider{}, nil
	}

	return &nullProvider{}, errors.New("Provider `" + name + "` is not supported.")
}
