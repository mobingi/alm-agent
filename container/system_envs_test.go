package container

import (
	"testing"

	"github.com/mobingi/alm-agent/config"
	"github.com/stretchr/testify/assert"
	"github.com/mobingi/alm-agent/metavars"
)

func TestStackID(t *testing.T) {
	tc := &config.Config{
		StackID: "teststack"
	}
	assert := assert.New(t)
	assert.Equal("STACK_ID=teststack", envFuncs.StackID(tc))
}

func InstanceID(t *testing.T) {
	metavars.ServerID = "dummyid"
	defer metavars.ServerID = ""
	tc := &config.Config{}
	assert := assert.New(t)
	assert.Equal("INSTANCE_ID=dummyid", envFuncs.InstanceID(tc))
}
