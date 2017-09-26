package container

import (
	"encoding/json"
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

func TestMackerelEnvs(t *testing.T) {
	defer func() { metavars.ServerID = "" }()

	assert := assert.New(t)
	tc := &config.Config{}
	tc.StackID = "testStack"
	tc.Flag = "testFlag"

	expect := []string{
		"apikey=dummykey",
		"auto_retirement=0",
		"enable_docker_plugin=1",
		"opts=-role testStack:testFlag",
	}

	js := `{"name": "mackarel", "apiKey": "dummykey"}`
	var opts interface{}
	json.Unmarshal([]byte(js), &opts)
	assert.Equal(expect, envFuncs.MackerelEnvs(tc, opts))
}
