package api

import (
	"testing"

	"github.com/mobingi/alm-agent/config"
)

type mockClient struct {
	client
}

func TestSetConfig(t *testing.T) {
	tc := &config.Config{
		APIHost: "https://test.example.com",
	}
	SetConfig(tc)
	if c.getConfig().APIHost != "https://test.example.com" {
		t.Fatal("Failed set Config.")
	}
}

func TestBuildURI(t *testing.T) {
	tc := &config.Config{
		APIHost: "https://test.example.com",
	}
	SetConfig(tc)

	expected := "https://test.example.com/v2/api"
	actual := c.buildURI("/v2/api")
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}
}
