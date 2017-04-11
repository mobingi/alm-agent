package main

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mobingilabs/go-modaemon/addons/aws/machine"
	"github.com/mobingilabs/go-modaemon/api"
	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/server_config"
)

func debug() bool {
	if os.Getenv("DEBUG") != "" {
		return true
	}
	return false
}

func init() {
	if debug() {
		log.SetLevel(log.DebugLevel)
		log.Debug("===== ENABLED Debug Logging =====")
	}

}

func main() {
	instance := machine.NewMachine()

	log.Debug("Step: config.LoadFromFile")
	moConfig, err := config.LoadFromFile("/opt/modaemon/modaemon.cfg")
	if err != nil {
		log.Debugf("%#v", err)
		log.Fatal("Failed load Config. exit.")
		os.Exit(1)
	}
	log.Debugf("%#v", moConfig)

	log.Debug("Step: api.NewClient")
	apiClient, err := api.NewClient(moConfig)
	if err != nil {
		log.Fatal("Failed load Config. exit.")
		os.Exit(1)
	}

	svConfig, err := apiClient.GetServerConfig(moConfig.APIHost)
	if err != nil {
		os.Exit(1)
	}
	log.Debugf("%#v", svConfig)

	creds := credentials.NewStaticCredentials(
		moConfig.AccessKey,
		moConfig.SecretKey,
		"",
	)

	awsconfig := aws.NewConfig().WithCredentials(creds).WithRegion(instance.Region)

	if debug() {
		awsconfig.WithLogLevel(aws.LogDebug)
	}

	sess := session.Must(session.NewSession(awsconfig))
	log.Debugf("%#v", sess)

	tmquit := make(chan bool)
	go func() {
		log.Debugf("Start checking inscale event.")
		if isTerminateWait(sess, instance) {
			log.Infof("Detected: Terminating:Wait")
			finalizeInstance(sess, instance, moConfig, svConfig)
			instance.SendLifeCycleAction(sess, moConfig, "CONTINUE")
			tmquit <- true
			return
		}
		tmcount := 0
		t := time.NewTicker(20 * time.Second)
		for {
			select {
			case <-t.C:
				log.Debugf("Start checking inscale event.")
				if isTerminateWait(sess, instance) {
					log.Infof("Detected: Terminating:Wait")
					finalizeInstance(sess, instance, moConfig, svConfig)
					instance.SendLifeCycleAction(sess, moConfig, "CONTINUE")
					tmquit <- true
					return
				}
			}
			tmcount++
			if tmcount == 2 {
				break
			}
		}
		t.Stop()
		tmquit <- true
	}()

	spquit := make(chan bool)
	go func() {
		if !instance.IsSpot {
			spquit <- true
			return
		}
		log.Debugf("Start checking spot termination event.")
		if instance.DetectSpotTerminationState() {
			log.Infof("Detected: Spot Instance Terminating")
			finalizeInstance(sess, instance, moConfig, svConfig)
			apiClient.SendSpotShutdownEvent(instance.InstanceID)
			spquit <- true
			return
		}
		spcount := 0
		t := time.NewTicker(15 * time.Second)
		for {
			select {
			case <-t.C:
				log.Debugf("Start checking spot termination event.")
				if instance.DetectSpotTerminationState() {
					log.Infof("Detected: Spot Instance Terminating")
					finalizeInstance(sess, instance, moConfig, svConfig)
					apiClient.SendSpotShutdownEvent(instance.InstanceID)
					spquit <- true
					return
				}
			}
			spcount++
			if spcount == 3 {
				break
			}
		}
		t.Stop()
		spquit <- true
	}()

	<-spquit
	<-tmquit
	os.Exit(0)
}

func isTerminateWait(sess *session.Session, instance *machine.Machine) bool {
	asState, err := instance.GetCurrentStateOfAS(sess)
	if err != nil {
		log.Debugf("%#v", err)
		return false
	}
	if asState != "Terminating:Wait" {
		return true
	}

	return false
}

func finalizeInstance(sess *session.Session, instance *machine.Machine, moConfig *config.Config, svConfig *serverConfig.Config) bool {
	instance.DeregisterInstancesFromELB(sess, moConfig)
	instance.CleanupCrontabs()
	instance.ExecShutdownTaskOnAppContainers(svConfig)

	asState, err := instance.GetCurrentStateOfAS(sess)
	if err != nil {
		log.Debugf("%#v", err)
		return false
	}
	if asState != "Terminating:Wait" {
		return true
	}

	return false
}
