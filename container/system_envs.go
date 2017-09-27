package container

import (
	dproxy "github.com/koron/go-dproxy"
	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/metavars"
	log "github.com/sirupsen/logrus"
)

// EnvHandle builds string to append to ENV.
type EnvHandle func(*config.Config, interface{}) []string

// EnvFuncs contains funcs of EnvHandle
type EnvFuncs struct{}

var envFuncs = &EnvFuncs{}

var handleEnvMap = map[string]EnvHandle{
	"stack_id":      envFuncs.StackID,
	"instance_id":   envFuncs.InstanceID,
	"mackerel_envs": envFuncs.MackerelEnvs,
}

// StackID resolves stack_id
func (e *EnvFuncs) StackID(c *config.Config, _ interface{}) []string {
	envs := []string{
		"STACK_ID=" + c.StackID,
	}
	return envs
}

// InstanceID resolves instance_id
func (e *EnvFuncs) InstanceID(c *config.Config, _ interface{}) []string {
	envs := []string{
		"INSTANCE_ID=" + metavars.ServerID,
	}
	return envs
}

// MackerelEnvs resolves mackerel_envs
func (e *EnvFuncs) MackerelEnvs(c *config.Config, opts interface{}) []string {
	op := dproxy.New(opts)
	log.Debugf("%#v", op)
	apikey, err := op.M("apiKey").String()
	if err != nil {
		log.Error("Addon: Faild to get Mackerel APIKEY.")
	}

	envs := []string{
		"apikey=" + apikey,
		"auto_retirement=0",
		"enable_docker_plugin=1",
		"opts=-v -role " + c.StackID + ":" + c.Flag,
	}
	return envs
}
