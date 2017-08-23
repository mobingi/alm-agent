package login

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mobingi/alm-agent/util"
)

func ExampleEnsureUser() {
	tmpHomeDir, _ := ioutil.TempDir("", "home")
	defer os.RemoveAll(tmpHomeDir)
	userHomeDir = tmpHomeDir

	util.Executer = &util.MockExecuter{}
	setLogin = func(username string, sshkey string) {
		return
	}
	EnsureUser("mobingi", "ssh-rsa PubKey")
	// Output:
	// useradd -m mobingi
	// usermod -aG docker mobingi
	// install -d -m 0700 -o mobingi -g mobingi WIP ...path.to
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
