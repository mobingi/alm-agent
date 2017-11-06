package main

import (
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mobingi/alm-agent/addons/aws/machine"
	"github.com/mobingi/alm-agent/api"
	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/server_config"
)

var (
	wg              sync.WaitGroup
	agentConfigPath = "/opt/mobingi/etc/alm-agent.cfg"
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
	agentConfig, err := config.LoadFromFile(agentConfigPath)
	if err != nil {
		log.Debugf("%#v", err)
		log.Fatal("Failed load Config. exit.")
		os.Exit(1)
	}
	log.Debugf("%#v", agentConfig)

	api.SetConfig(agentConfig)
	err = api.GetAccessToken()
	if err != nil {
		log.Fatal("Failed to Get Access Token from API")
		os.Exit(1)
	}

	svConfig, err := api.GetServerConfig(agentConfig.APIHost)
	if err != nil {
		os.Exit(1)
	}
	log.Debugf("%#v", svConfig)

	// use EC2 InstanceRole
	sess := session.Must(session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigDisable,
		}))

	metasvc := ec2metadata.New(sess)
	region, err := metasvc.Region()
	if err != nil {
		log.Fatal("Faild to detect region.")
		os.Exit(1)
	}

	awsconfig := &aws.Config{
		Region: aws.String(region),
	}
	// new session with awsconfig
	sess = session.Must(session.NewSessionWithOptions(
		session.Options{
			Config:            *awsconfig,
			SharedConfigState: session.SharedConfigDisable,
		}))

	if debug() {
		awsconfig.WithLogLevel(aws.LogDebug)
	}

	log.Debugf("%#v", sess)

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Debugf("Start checking inscale event.")
		// Run first check
		if isTerminateWait(sess, instance) {
			finalizeNormalInstance(sess, instance, agentConfig, svConfig)
			return
		}
		t := time.NewTicker(20 * time.Second)
		defer t.Stop()
		for tmcount := 0; tmcount < 2; tmcount++ {
			select {
			case <-t.C:
				log.Debugf("Start checking inscale event.")
				if isTerminateWait(sess, instance) {
					finalizeNormalInstance(sess, instance, agentConfig, svConfig)
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
			finalizeSpotInstance(sess, instance, agentConfig, svConfig)
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
					finalizeSpotInstance(sess, instance, agentConfig, svConfig)
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
	switch asState {
	case "":
		log.Debug("Not in ASG.")
		os.Exit(0)
	case "Terminating:Wait":
		return true
	}

	return false
}

func finalizeInstance(sess *session.Session, instance *machine.Machine, agentConfig *config.Config, svConfig *serverConfig.Config) {
	instance.DeregisterInstancesFromELB(sess, agentConfig)
	instance.CleanupCrontabs()
	instance.ExecShutdownTaskOnAppContainers(agentConfig, svConfig)
	return
}

func finalizeNormalInstance(sess *session.Session, instance *machine.Machine, agentConfig *config.Config, svConfig *serverConfig.Config) {
	log.Infof("Detected: Terminating:Wait")
	finalizeInstance(sess, instance, agentConfig, svConfig)
	instance.SendLifeCycleAction(sess, agentConfig, "CONTINUE")
	return
}

func finalizeSpotInstance(sess *session.Session, instance *machine.Machine, agentConfig *config.Config, svConfig *serverConfig.Config) {
	log.Infof("Detected: Spot Instance Terminating")
	finalizeInstance(sess, instance, agentConfig, svConfig)
	return
}
