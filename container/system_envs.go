package container

import (
	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/metavars"
)

// EnvHandle builds string to append to ENV.
type EnvHandle func(*config.Config) string

// EnvFuncs contains funcs of EnvHandle
type EnvFuncs struct{}

var envFuncs = &EnvFuncs{}

var handleEnvMap = map[string]EnvHandle{
	"stack_id":    envFuncs.StackID,
	"instance_id": envFuncs.InstanceID,
}

// StackID resolves stack_id
func (e *EnvFuncs) StackID(c *config.Config) string {
	return "STACK_ID=" + c.StackID
}

// InstanceID resolves instance_id
func (e *EnvFuncs) InstanceID(c *config.Config) string {
	return "INSTANCE_ID=" + metavars.ServerID
}
