package container

import "net"

// Container means docker container
type Container struct {
	Name string
	ID   string
	IP   net.IP
}
