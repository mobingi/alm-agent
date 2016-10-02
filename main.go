package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"

	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/container/docker"
	"github.com/mobingilabs/go-modaemon/server_config"
)

func main() {
	const defaultModaemonConfig = "/opt/modaemon/modaemon.cfg"

	var modaemonConfig string
	flag.StringVar(&modaemonConfig, "c", defaultModaemonConfig, "path of modaemon.cfg")
	flag.Parse()

	c, err := config.LoadFromFile(modaemonConfig)

	if err != nil {
		log.Fatal(err)
	}

	s, err := serverConfig.Get(c.ServerConfigAPIEndPoint)

	d, err := docker.New(s.Image, s.DockerHubUserName, s.DockerHubPassword)

	if err != nil {
		log.Fatal(err)
	}

	err = d.ImagePull()
	if err != nil {
		log.Fatal(err)
	}
}
