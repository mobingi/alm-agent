package machine

import (
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/mobingi/alm-agent/config"
)

// GetCurrentStateOfAS returns Instance State on AutoScalling
// eg. InService, Terminating:Wait
func (m *Machine) GetCurrentStateOfAS(sess *session.Session) (string, error) {
	asClient := autoscaling.New(sess)
	asparams := &autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []*string{
			aws.String(m.InstanceID),
		},
	}

	asresp, err := asClient.DescribeAutoScalingInstances(asparams)
	if err != nil {
		log.Debugf("%#v", err)
		return "UNKONWN", err
	}

	log.Debugf("%#v", asresp)
	if len(asresp.AutoScalingInstances) == 0 {
		return "", nil
	}
	m.ASName = *asresp.AutoScalingInstances[0].AutoScalingGroupName
	return *asresp.AutoScalingInstances[0].LifecycleState, nil
}

// DeregisterInstancesFromELB removes instance from ELB backend servers.
func (m *Machine) DeregisterInstancesFromELB(sess *session.Session, agentConfig *config.Config) {
	cfnClient := cloudformation.New(sess)

	cfnparams := &cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(agentConfig.StackID),
	}

	cfnresp, err := cfnClient.DescribeStackResources(cfnparams)
	if err != nil {
		log.Debugf("%#v", err)
		return
	}
	// log.Debugf("%#v", cfnresp)

	myelbID := ""
	for _, x := range cfnresp.StackResources {
		if *x.ResourceType == "AWS::ElasticLoadBalancing::LoadBalancer" {
			myelbID = *x.PhysicalResourceId
			break
		}
	}
	log.Debugf("myelbID: %#v", myelbID)

	if myelbID == "" {
		return
	}

	elbClient := elb.New(sess)
	elbparams := &elb.DeregisterInstancesFromLoadBalancerInput{
		Instances: []*elb.Instance{
			{
				InstanceId: aws.String(m.InstanceID),
			},
		},
		LoadBalancerName: aws.String(myelbID),
	}
	elbresp, err := elbClient.DeregisterInstancesFromLoadBalancer(elbparams)
	if err != nil {
		log.Debugf("%#v", err)
		return
	}
	log.Debugf("%#v", elbresp)
	log.Info("Instance deregistered from ELB.")
	return
}

// SendLifeCycleAction tell dying to API
func (m *Machine) SendLifeCycleAction(sess *session.Session, agentConfig *config.Config, action string) bool {
	asClient := autoscaling.New(sess)

	asparams := &autoscaling.CompleteLifecycleActionInput{
		AutoScalingGroupName:  aws.String(m.ASName),
		LifecycleActionResult: aws.String(action),
		LifecycleHookName:     aws.String(agentConfig.StackID + "-lch-ists-tmnt"),
		InstanceId:            aws.String(m.InstanceID),
	}

	asresp, err := asClient.CompleteLifecycleAction(asparams)
	if err != nil {
		log.Debugf("%#v", err)
		return false
	}

	log.Debugf("%#v", asresp)
	return true
}
