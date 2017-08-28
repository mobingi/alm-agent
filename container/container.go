package container

import "net"

// Container means docker container
type Container struct {
	Name  string
	ID    string
	IP    net.IP
	State string
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
