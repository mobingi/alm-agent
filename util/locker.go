package util

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var lockbase = "/var/lock"
var lockfile = func(name string) string {
	return filepath.Join(lockbase, name+".lock")
}

var lockmap = map[string]int{}

// Lock creates flock based lockfile
func Lock(name string) error {
	log.Debugf("try util.Lock %s", name)
	if _, err := os.Stat(lockfile(name)); err != nil {
		log.Debug(err)
		if err := ioutil.WriteFile(lockfile(name), []byte(""), 0644); err != nil {
			log.Error(err)
			return errors.New("gpp")
		}
	}
	fd, _ := syscall.Open(lockfile(name), syscall.O_RDONLY, 0000)
	lockmap[name] = fd

	if err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		log.Warn(err)
		return errors.New("resource temporarily unavailable")
	}

	log.Debugf("Locked %s", lockmap)
	return nil
}

// UnLock clears flock based lockfile
// should call UnLock by defer
func UnLock(name string) error {
	log.Debugf("try util.UnLock %s", name)
	syscall.Close(lockmap[name])

	delete(lockmap, name)
	log.Debugf("UnLocked %s", lockmap)
	return nil
}
