package util

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLock(t *testing.T) {
	assert := assert.New(t)

	tmpLogDir, _ := ioutil.TempDir("", "lockdir")
	defer os.RemoveAll(tmpLogDir)

	origlockbase := lockbase
	lockbase = tmpLogDir
	defer func() {
		lockbase = origlockbase
		lockmap = map[string]int{}
	}()

	defer UnLock("testlock")
	assert.NoError(Lock("testlock"))
	assert.NoError(UnLock("testlock"))
	assert.NoError(Lock("testlock"))
	// twice
	assert.Error(Lock("testlock"))
}

func TestUnLock(t *testing.T) {
	assert := assert.New(t)

	tmpLogDir, _ := ioutil.TempDir("", "lockdir")
	defer os.RemoveAll(tmpLogDir)

	origlockbase := lockbase
	lockbase = tmpLogDir
	defer func() {
		lockbase = origlockbase
		lockmap = map[string]int{}
	}()

	defer UnLock("testlock")
	assert.False(FileExists(lockfile("testlock")))
	Lock("testlock")
	assert.True(FileExists(lockfile("testlock")))

	// lock after unlocked
	assert.NoError(UnLock("testlock"))
	assert.NoError(Lock("testlock"))
}
