package login

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mobingi/alm-agent/util"
)

func TestEnsureUser(t *testing.T) {
	defer util.ClearMockBuffer()
	tmpHomeDir, _ := ioutil.TempDir("", "home")
	defer os.RemoveAll(tmpHomeDir)

	origuserHomeDir := userHomeDir
	userHomeDir = tmpHomeDir
	defer func() { userHomeDir = origuserHomeDir }()

	orig_setLogin := setLogin
	defer func() { setLogin = orig_setLogin }()
	setLogin = func(username string, sshkey string) {
		return
	}

	util.Executer = &util.MockExecuter{}
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

}

func TestSshDirpath(t *testing.T) {
	expected := "/home/mobingi/.ssh"
	actual := sshDirpath("mobingi")
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}

	origuserHomeDir := userHomeDir
	userHomeDir = "/tmp"
	defer func() { userHomeDir = origuserHomeDir }()

	expected = "/tmp/mobingi/.ssh"
	actual = sshDirpath("mobingi")
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}
}
