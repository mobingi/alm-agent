package container

import (
	"testing"

	"github.com/mobingi/alm-agent/config"
	"github.com/stretchr/testify/assert"
)

func TestStackID(t *testing.T) {
	tc := &config.Config{}
	assert := assert.New(t)
	assert.Equal("STACK_ID=", envFuncs.StackID(tc))
}
