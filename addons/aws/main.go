package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mobingilabs/go-modaemon/addons/aws/machine"
	"github.com/mobingilabs/go-modaemon/api"
	"github.com/mobingilabs/go-modaemon/config"
)

func main() {
	debug := os.Getenv("DEBUG")
	if debug != "" {
		log.SetLevel(log.DebugLevel)
	}

	log.Debug("Step: config.LoadFromFile")

	instance := machine.NewMachine()
	fmt.Printf("%#v\n", instance)

	conf, err := config.LoadFromFile("/opt/modaemon/modaemon.cfg")
	if err != nil {
		log.Debugf("%#v", err)
		os.Exit(1)
	}
	log.Debugf("%#v", conf)

	log.Debug("Step: api.NewClient")
	apiClient, err := api.NewClient(conf)
	if err != nil {
		os.Exit(1)
	}

	s, err := apiClient.GetServerConfig(conf.APIHost)
	if err != nil {
		os.Exit(1)
	}
	log.Debugf("%#v", s)

	t, err := apiClient.GetStsToken()
	if err != nil {
		os.Exit(1)
	}
	log.Debugf("%#v", t)

	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String(instance.Region),
		Credentials: credentials.NewStaticCredentials(t.AccessKeyID, t.SecretAccessKey, t.SessionToken),
	})

	log.Debugf("%#v", sess)
}
