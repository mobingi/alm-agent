package util

import (
	"io/ioutil"
	"os"
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
