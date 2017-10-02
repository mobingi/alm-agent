package container

import (
	"context"
	"net"

	"docker.io/go-docker/api/types/container"
	log "github.com/sirupsen/logrus"
)

// Container means docker container
type Container struct {
	Name  string
	ID    string
	IP    net.IP
	State string
}

// CompareToRestart used to check to need restart.
func CompareToRestart(name string, newDocker *Docker, currentInfo *Container, addcon *SystemContainer) {
	currentCT, err := newDocker.Client.ContainerInspect(context.Background(), currentInfo.ID)
	if err != nil {
		addcon.Restart = true
		return
	}

	log.Debugf("newDocker: %#v", newDocker)
	log.Debugf("currentCT.Config: %#v", currentCT.Config)
	addcon.Restart = handleCompareMap[name](newDocker, currentCT.Config)
	return
}

// CompareHandle retruns bool to set addcon.Restart.
type CompareHandle func(*Docker, *container.Config) bool

// CompareFuncs contains funcs of CompareHandle by name
type CompareFuncs struct{}

var compareFuncs = &CompareFuncs{}

var handleCompareMap = map[string]CompareHandle{
	"mackerel": compareFuncs.Mackerel,
}

// Mackerel resolves mackerel
// should detect change of apiKey
func (c *CompareFuncs) Mackerel(dc *Docker, conf *container.Config) bool {
	i := 0
	for _, x := range dc.Envs {
		for _, y := range conf.Env {
			if x == y {
				i++
			}
		}
	}

	return len(dc.Envs) != i
}

// State is state of Container from Docker API.
/*
  "State": {
    "Status": "exited",
    "Running": false,
    "Paused": false,
    "Restarting": false,
    "OOMKilled": false,
    "Dead": false,
    "Pid": 0,
    "ExitCode": 0,
    "Error": "",
    "StartedAt": "2017-08-27T05:34:21.077356196Z",
    "FinishedAt": "2017-08-27T05:34:22.564500324Z"
	},
*/
// type State struct {
// 	Status   string
// 	Running  bool
// 	ExitCode int
// 	Error    string
// }
