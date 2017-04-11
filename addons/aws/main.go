package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mobingilabs/go-modaemon/addons/aws/machine"
	"github.com/mobingilabs/go-modaemon/api"
	"github.com/mobingilabs/go-modaemon/config"
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

	asState := instance.GetCurrentStateOfAS(sess)

	if asState != "Terminating:Wait" {
		return
	}

	log.Infof("Detected: Terminating:Wait")
	instance.DeregisterInstancesFromELB(sess, moConfig)
	instance.CleanupCrontabs()
	instance.SendLifeCycleAction(sess, moConfig, "CONTINUE") // InScale Only

	// TODO: exec_shutdown_tasks_on_app_containers => Machine
	// TODO: send_notification_to_api // SPOT Only
}
