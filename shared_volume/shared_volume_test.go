package sharedvolume

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullVolume(t *testing.T) {
	assert := assert.New(t)
	vol := &NullVolume{}

	assert.Nil(vol.Setup())
	assert.Nil(vol.load())
}
