package api

import (
	"testing"

	"github.com/mobingi/alm-agent/config"
)

func TestSetConfig(t *testing.T) {
	tc := &config.Config{
		APIHost: "test.example.com",
	}
	SetConfig(tc)
	if c.getConfig().APIHost != "test.example.com" {
		t.Error("Failed set Config.")
	}
}
