package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	tmpFile, _ := ioutil.TempFile("", "tmptest")
	defer os.Remove(tmpFile.Name())

	if !FileExists(tmpFile.Name()) {
		t.Error("File exists but returns false.")
	}

	if FileExists("/tmp/path/to/not/exists") {
		t.Error("File does not exists but returns true.")
	}
}

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
