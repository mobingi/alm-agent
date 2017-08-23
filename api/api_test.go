package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestGetBase(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("Expected method %s; got %q", http.MethodGet, r.Method)
		}
		fmt.Fprintln(w, "{}")
	}))
	defer ts.Close()

	t.Log(ts.URL)
	tc := &config.Config{
		APIHost: ts.URL,
	}
	SetConfig(tc)

	values := url.Values{}
	err := Get("/testget", values, nil)
	if err != nil {
		t.Fatalf("Error on Get Request: %#v", err)
	}
}

func TestPostBase(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected method %s; got %q", http.MethodPost, r.Method)
		}
		fmt.Fprintln(w, "{}")
	}))
	defer ts.Close()

	t.Log(ts.URL)
	tc := &config.Config{
		APIHost: ts.URL,
	}
	SetConfig(tc)

	values := url.Values{}
	err := Post("/testpost", values, nil)
	if err != nil {
		t.Fatalf("Error on Get Request: %#v", err)
	}
}
