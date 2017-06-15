package login

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingilabs/go-modaemon/util"
)

func EnsureUser(username string, sshkey string) {
	log.Debug("Step: ensureUser")

	_, err := user.Lookup(username)
	if err != nil {
		log.Infof("Step: Useradd %s\n", username)
		exec.Command("useradd", "-m", username).Run()
		exec.Command("usermod", "-aG", "docker", username).Run()
		exec.Command("install", "-d", "-m", "0700", "-o", username, "-g", username, sshDirpath(username)).Run()
	} else {
		log.Debugf("User %s already exists.\n", username)
	}

	if !util.FileExists(filepath.Join(sshDirpath(username), "authorized_keys")) {
		setLogin(username, sshkey)
	} else if !checkIsKeySame(username, sshkey) {
		setLogin(username, sshkey)
	} else {
		log.Debugf("No need update sshkey for %s.\n", username)
	}

	return
}

func sshDirpath(username string) string {
	return filepath.Join("/home", username, ".ssh")
}

func checkIsKeySame(username string, sshkey string) bool {
	dat, _ := ioutil.ReadFile(filepath.Join(sshDirpath(username), "authorized_keys"))

	return strings.Contains(string(dat), sshkey)
}

func setLogin(username string, sshkey string) {
	log.Debugf("Step: setLogin for %s.\n", username)
	content := fmt.Sprintf("command=\"docker exec -t -i active /bin/bash\" %s", sshkey)
	err := ioutil.WriteFile(filepath.Join(sshDirpath(username), "authorized_keys"), []byte(content), 0600)
	if err != nil {
		log.Errorf("Faild adding user: %s\n", username)
	}
	currentUser, _ := user.Lookup(username)
	uid, _ := strconv.Atoi(currentUser.Uid)
	gid, _ := strconv.Atoi(currentUser.Gid)
	syscall.Chown(filepath.Join(sshDirpath(username), "authorized_keys"), uid, gid)
}
