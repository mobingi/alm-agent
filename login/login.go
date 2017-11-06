package login

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/mobingi/alm-agent/util"
)

var userHomeDir = "/home"

var protectedUsers = []string{
	"adm",
	"bin",
	"daemon",
	"dbus",
	"ec2-user",
	"ftp",
	"games",
	"gopher",
	"halt",
	"k5user",
	"lp",
	"mail",
	"mailnull",
	"nfsnobody",
	"nobody",
	"ntp",
	"operator",
	"root",
	"rpc",
	"rpcuser",
	"saslauth",
	"shutdown",
	"smmsp",
	"sshd",
	"sync",
	"uucp",
}

// EnsureUser stores publickey for users
func EnsureUser(username string, sshkey string) {

	log.Debug("Step: ensureUser")

	for _, u := range protectedUsers {
		if u == username {
			log.Warnf("%s is protected user. Setup skipped.\n", username)
			return
		}
	}

	_, err := user.Lookup(username)
	if err != nil {
		log.Infof("Step: Useradd %s\n", username)
		util.Executor.Exec("useradd", "-m", username)
		util.Executor.Exec("usermod", "-aG", "docker", username)
		util.Executor.Exec("install", "-d", "-m", "0700", "-o", username, "-g", username, sshDirpath(username))
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
	return filepath.Join(userHomeDir, username, ".ssh")
}

func checkIsKeySame(username string, sshkey string) bool {
	dat, _ := ioutil.ReadFile(filepath.Join(sshDirpath(username), "authorized_keys"))

	return strings.Contains(string(dat), sshkey)
}

var setLogin = func(username string, sshkey string) {
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
	return
}
