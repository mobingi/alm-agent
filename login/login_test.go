package login

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mobingi/alm-agent/util"
)

func TestEnsureUser(t *testing.T) {
	util.ClearMockBuffer()

	tmpHomeDir, _ := ioutil.TempDir("", "home")
	defer os.RemoveAll(tmpHomeDir)
	userHomeDir = tmpHomeDir

	util.Executer = &util.MockExecuter{}
	setLogin = func(username string, sshkey string) {
		return
	}
	EnsureUser("mobingi", "ssh-rsa PubKey")
	buf := util.GetMockBuffer()
	t.Log(buf)
	expected := "useradd -m mobingi"
	actual := buf[0]
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}

	expected = "usermod -aG docker mobingi"
	actual = buf[1]
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}

	expected = "install -d -m 0700 -o mobingi -g mobingi " + tmpHomeDir + "/mobingi/.ssh"
	actual = buf[2]
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}
	util.ClearMockBuffer()
}

func TestSshDirpath(t *testing.T) {
	expected := "/home/mobingi/.ssh"
	actual := sshDirpath("mobingi")
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}

	userHomeDir = "/tmp"
	expected = "/tmp/mobingi/.ssh"
	actual = sshDirpath("mobingi")
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}
}
