package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchContainerState(t *testing.T) {
	tmpLogDir, _ := ioutil.TempDir("", "containerLogs")
	containerLogsLocation = tmpLogDir
	defer os.RemoveAll(tmpLogDir)

	os.Mkdir(filepath.Join(tmpLogDir, "log"), 0755)

	if FetchContainerState() != "" {
		t.Error("container_status not Nil!")
	}

	ioutil.WriteFile(filepath.Join(tmpLogDir, "log", "container_status"), []byte("running"), 0644)
	if FetchContainerState() != "running" {
		t.Error("FetchContainerState Failed!")
	}
}
