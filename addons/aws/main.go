package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
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

	cred := credentials.NewStaticCredentials(
		t.AccessKeyID,
		t.SecretAccessKey,
		t.SessionToken,
	)
	awsconfig := aws.NewConfig().WithCredentials(cred).WithRegion(instance.Region)

	if debug != "" {
		awsconfig.WithLogLevel(aws.LogDebug)
	}

	sess := session.Must(session.NewSession(awsconfig))
	// sess := session.Must(session.NewSession(aws.NewConfig().WithLogLevel(aws.LogDebug)))
	log.Debugf("%#v", sess)

	as := autoscaling.New(sess)
	asparams := &autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []*string{
			aws.String(instance.InstanceID),
		},
	}

	asresp, err := as.DescribeAutoScalingInstances(asparams)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		log.Fatalf("%#v", err)
		return
	}

	log.Debugf("%#v", asresp)
	log.Infof("State: %s", *asresp.AutoScalingInstances[0].LifecycleState)

	if *asresp.AutoScalingInstances[0].LifecycleState != "Terminating:Wait" {
		return
	}

	log.Infof("Detected: Terminating:Wait")
	// TODO: SPOT: deregister_instances_from_load_balancer => Machine
	// TODO: cleanup crontab => Machine
	// TODO: exec_shutdown_tasks_on_app_containers => Machine
	// TODO: complete_lifecycle_action `CONTINUE``
	// TODO: send_notification_to_api
}
