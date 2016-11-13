package container

import "net"

type Container struct {
	Name string
	ID   string
	IP   net.IP
}
