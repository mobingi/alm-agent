package container

import (
	"testing"

	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/metavars"
	"github.com/stretchr/testify/assert"
)

func TestStackID(t *testing.T) {
	tc := &config.Config{
		StackID: "teststack",
	}
	assert := assert.New(t)
	assert.Equal([]string{"STACK_ID=teststack"}, envFuncs.StackID(tc, nil))
}

func TestInstanceID(t *testing.T) {
	metavars.ServerID = "dummyid"
	defer func() { metavars.ServerID = "" }()
	tc := &config.Config{}
	assert := assert.New(t)
	assert.Equal([]string{"INSTANCE_ID=dummyid"}, envFuncs.InstanceID(tc, nil))
}
