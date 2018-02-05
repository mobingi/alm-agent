package api

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mobingi/alm-agent/util"
	"github.com/stretchr/testify/assert"
)

func Test_saveStatus(t *testing.T) {
	assert := assert.New(t)

	tmpMobDir, _ := ioutil.TempDir("", "TestStatus")
	defer os.RemoveAll(tmpMobDir)

	origlastAgentStatusPath := lastAgentStatusPath
	lastAgentStatusPath = filepath.Join(tmpMobDir, "last_agent_status")
	defer func() { lastAgentStatusPath = origlastAgentStatusPath }()

	origlastContainerStatusPath := lastContainerStatusPath
	lastContainerStatusPath = filepath.Join(tmpMobDir, "last_agent_status")
	defer func() { lastContainerStatusPath = origlastContainerStatusPath }()

	saveStatus(lastAgentStatusPath, "testinga")
	assert.True(util.FileExists(lastAgentStatusPath))

	saveStatus(lastContainerStatusPath, "testingc")
	assert.True(util.FileExists(lastContainerStatusPath))
}

func Test_isNewStatus(t *testing.T) {
	assert := assert.New(t)

	tmpMobDir, _ := ioutil.TempDir("", "TestStatus")
	defer os.RemoveAll(tmpMobDir)

	origlastAgentStatusPath := lastAgentStatusPath
	lastAgentStatusPath = filepath.Join(tmpMobDir, "last_agent_status")
	defer func() { lastAgentStatusPath = origlastAgentStatusPath }()

	origlastContainerStatusPath := lastContainerStatusPath
	lastContainerStatusPath = filepath.Join(tmpMobDir, "last_agent_status")
	defer func() { lastContainerStatusPath = origlastContainerStatusPath }()

	saveStatus(lastAgentStatusPath, "testinga")
	assert.False(isNewStatus(lastAgentStatusPath, "testinga"))
	assert.True(isNewStatus(lastAgentStatusPath, "testinger"))

	saveStatus(lastContainerStatusPath, "testingc")
	assert.False(isNewStatus(lastContainerStatusPath, "testingc"))
	assert.True(isNewStatus(lastContainerStatusPath, "testinger"))
}
