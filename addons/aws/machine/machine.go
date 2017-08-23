package machine

import (
	"io/ioutil"
	"net/http"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

// METAENDPOINT EC2 Metadata Endpoint
// we can update machine.METAENDPOINT on build.
var METAENDPOINT = "http://169.254.169.254/"
var instanceLifecycle = "/opt/mobingi/etc/instance_lifecycle"

var getInstanceID = func(m *Machine) string {
	resp, err := http.Get(METAENDPOINT + "/latest/meta-data/instance-id")
	if err != nil {
		log.Fatalf("%#v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("%#v", err)
	}
	return string(body)
}

var getRegion = func(m *Machine) string {
	resp, err := http.Get(METAENDPOINT + "/latest/meta-data/placement/availability-zone")
	if err != nil {
		log.Fatalf("%#v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("%#v", err)
	}
	az := string(body)
	return az[0:(len(az) - 1)]
}

// Machine means EC2 Insatnce.
type Machine struct {
	InstanceID string
	Region     string
	IsSpot     bool
	ASName     string
}

// NewMachine as constructor.
func NewMachine() *Machine {
	machine := new(Machine)
	machine.InstanceID = getInstanceID(machine)
	machine.Region = getRegion(machine)
	machine.IsSpot = isSpot()
	return machine
}

func isSpot() bool {
	dat, err := ioutil.ReadFile(instanceLifecycle)

	if err != nil {
		// for older template
		// I will regard it as True to run checker.
		return true
	}

	if string(dat) == "spot" {
		return true
	}

	return false
}

// DetectSpotTerminationState returns Spot Instance is terminating or not.
func (m *Machine) DetectSpotTerminationState() bool {
	resp, err := http.Get(METAENDPOINT + "/latest/meta-data/spot/termination-time")
	if err != nil {
		log.Debugf("%#v", err)
		return false
	}

	log.Debugf("%d", resp.StatusCode)
	if resp.StatusCode != 404 {
		return true
	}

	return false
}

// CleanupCrontabs removes all jobs.
func (m *Machine) CleanupCrontabs() bool {
	exec.Command("crontab -r -u root").Run()
	return false
}
