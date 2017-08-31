package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mobingi/alm-agent/metavars"
	"github.com/stretchr/testify/assert"
)

func TestFetchContainerState(t *testing.T) {
	assert := assert.New(t)

	tmpLogDir, _ := ioutil.TempDir("", "containerLogs")
	defer os.RemoveAll(tmpLogDir)

	oringcontainerLogsLocation := containerLogsLocation
	containerLogsLocation = tmpLogDir
	defer func() { containerLogsLocation = oringcontainerLogsLocation }()

	os.Mkdir(filepath.Join(tmpLogDir, "log"), 0755)

	assert.Empty(FetchContainerState())

	ioutil.WriteFile(filepath.Join(tmpLogDir, "log", "container_status"), []byte("running"), 0644)
	assert.Equal("running", FetchContainerState())
}

func TestAgentID(t *testing.T) {
	assert := assert.New(t)
	tmpFile, _ := ioutil.TempFile("", "alm-agent.id")
	defer os.Remove(tmpFile.Name())
	origagentIDSavePath := agentIDSavePath
	agentIDSavePath = tmpFile.Name()
	defer func() {
		agentIDSavePath = origagentIDSavePath
		metavars.AgentID = ""
	}()

	// Create New ID
	AgentID()

	dat, _ := ioutil.ReadFile(agentIDSavePath)
	assert.Equal(string(dat), metavars.AgentID)

	// Read From File
	tmpID := "alm-agent-tmp-id"
	ioutil.WriteFile(agentIDSavePath, []byte(tmpID), 0644)
	AgentID()
	assert.Equal(string(tmpID), metavars.AgentID)
}
