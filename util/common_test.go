package util

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	assert := assert.New(t)
	tmpFile, _ := ioutil.TempFile("", "tmptest")
	defer os.Remove(tmpFile.Name())

	assert.True(FileExists(tmpFile.Name()))
	assert.False(FileExists("/tmp/path/to/not/exists"))
}
