package main

import (
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mobingi/alm-agent/addons/aws/machine"
	"github.com/mobingi/alm-agent/api"
	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/server_config"
)

var wg sync.WaitGroup

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

	svConfig, err := api.GetServerConfig(moConfig.APIHost)
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Debugf("Start checking inscale event.")
		// Run first check
		if isTerminateWait(sess, instance) {
			finalizeNormalInstance(sess, instance, moConfig, svConfig)
			return
		}
		t := time.NewTicker(20 * time.Second)
		defer t.Stop()
		for tmcount := 0; tmcount < 2; tmcount++ {
			select {
			case <-t.C:
				log.Debugf("Start checking inscale event.")
				if isTerminateWait(sess, instance) {
					finalizeNormalInstance(sess, instance, moConfig, svConfig)
					return
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if !instance.IsSpot {
			return
		}
		log.Debugf("Start checking spot termination event.")
		// Run first check
		if instance.DetectSpotTerminationState() {
			finalizeSpotInstance(sess, instance, moConfig, svConfig)
			api.SendSpotShutdownEvent(instance.InstanceID)
			return
		}
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		for spcount := 0; spcount < 3; spcount++ {
			select {
			case <-t.C:
				log.Debugf("Start checking spot termination event.")
				if instance.DetectSpotTerminationState() {
					finalizeSpotInstance(sess, instance, moConfig, svConfig)
					api.SendSpotShutdownEvent(instance.InstanceID)
					return
				}
			}
		}
	}()

	wg.Wait()
	os.Exit(0)
}

func isTerminateWait(sess *session.Session, instance *machine.Machine) bool {
	asState, err := instance.GetCurrentStateOfAS(sess)
	if err != nil {
		log.Debugf("%#v", err)
		return false
	}
	if asState == "Terminating:Wait" {
		return true
	}

	return false
}

func finalizeInstance(sess *session.Session, instance *machine.Machine, moConfig *config.Config, svConfig *serverConfig.Config) {
	instance.DeregisterInstancesFromELB(sess, moConfig)
	instance.CleanupCrontabs()
	instance.ExecShutdownTaskOnAppContainers(moConfig, svConfig)
	return
}

func finalizeNormalInstance(sess *session.Session, instance *machine.Machine, moConfig *config.Config, svConfig *serverConfig.Config) {
	log.Infof("Detected: Terminating:Wait")
	finalizeInstance(sess, instance, moConfig, svConfig)
	instance.SendLifeCycleAction(sess, moConfig, "CONTINUE")
	return
}

func finalizeSpotInstance(sess *session.Session, instance *machine.Machine, moConfig *config.Config, svConfig *serverConfig.Config) {
	log.Infof("Detected: Spot Instance Terminating")
	finalizeInstance(sess, instance, moConfig, svConfig)
	return
}
