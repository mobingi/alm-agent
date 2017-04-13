package machine

import "testing"

func TestNewMachine(t *testing.T) {
	getInstanceID = func(m *Machine) string {
		return "i-XXXXXXXXXX"
	}
	getRegion = func(m *Machine) string {
		return "ap-northeast-1"
	}

	instance := NewMachine()

	if instance.InstanceID != "i-XXXXXXXXXX" {
		t.Fatalf("InstanceID Broken.")
	}

	if instance.Region != "ap-northeast-1" {
		t.Fatalf("Region Broken.")
	}

	if instance.IsSpot == false {
		t.Fatalf("Spot checker Broken.")
	}
}
