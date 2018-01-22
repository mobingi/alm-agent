package api

import (
	"testing"

	"github.com/mobingi/alm-agent/server_config"
	"github.com/stretchr/testify/assert"
)

func TestGetServerConfigFromFile(t *testing.T) {
	assert := assert.New(t)
	sc := &serverConfig.Config{}
	getServerConfigFromFile("../test/fixtures/serverconfig.v2.json", sc)

	assert.Equal(sc.Image, "mobingi/ubuntu-apache2-php7:7.1")
	assert.NotEmpty(sc.EnvironmentVariables)
}
