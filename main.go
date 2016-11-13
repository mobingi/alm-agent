package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"

	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/container"
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

	if err != nil {
		log.Fatal(err)
	}

	d, err := container.NewDocker(s)

	if err != nil {
		log.Fatal(err)
	}

	greenContainer, err := d.StartContainer("green", "/tmp")
	if err != nil {
		log.Fatal(err)
	}

	err = d.MapPort(greenContainer)
	if err != nil {
		log.Error(err)
	}

	err = d.UnmapPort(greenContainer)
	if err != nil {
		log.Error(err)
	}

	d.StopContainer(greenContainer)
	d.RemoveContainer(greenContainer)
}
