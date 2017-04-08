package machine

import (
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

var METAENDPOINT = "http://169.254.169.254/"

type Machine struct {
	InstanceID string
	Region     string
}

func NewMachine() *Machine {
	machine := new(Machine)
	machine.InstanceID = getInstanceID(machine)
	machine.Region = getRegion(machine)
	return machine
}

func getInstanceID(m *Machine) string {
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

func getRegion(m *Machine) string {
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
