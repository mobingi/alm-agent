package serverConfig

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNeedsUpdate(t *testing.T) {
	tc := &Config{}
	tc.Updated = uint(1500000000)

	tmpFile, _ := ioutil.TempFile("", "configVersion")
	defer os.Remove(tmpFile.Name())
	origversionPath := versionPath
	versionPath = tmpFile.Name()
	defer func() { versionPath = origversionPath }()

	// empty configVersion
	assert := assert.New(t)
	actual, _ := NeedsUpdate(tc)
	assert.True(actual)

	// local = server
	WriteUpdated(tc)
	actual, _ = NeedsUpdate(tc)
	assert.False(actual)

	// local < server
	tc.Updated = uint(1500000001)
	actual, _ = NeedsUpdate(tc)
	assert.True(actual)

	// local > server
	tc.Updated = uint(20000000)
	actual, _ = NeedsUpdate(tc)
	assert.False(actual)
}
