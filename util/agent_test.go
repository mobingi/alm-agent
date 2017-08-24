package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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
