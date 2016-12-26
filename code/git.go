package code

import (
	"fmt"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

type Git struct {
	url  string
	path string
	ref  string
}

func (g *Git) checkUpdate() (bool, error) {
	cmd := exec.Command("git", "fetch")
	cmd.Dir = g.path
	err := cmd.Run()
	if err != nil {
		return false, err
	}

	cmd = exec.Command("git", "diff", fmt.Sprintf("origin/%s", g.ref))
	cmd.Dir = g.path

	out, err := cmd.Output()

	if err != nil {
		return false, err
	}

	if len(out) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (g *Git) get() error {
	log.Infof("Executing git clone -b %s %s %s", g.ref, g.url, g.path)
	out, err := exec.Command("git", "clone", "-b", g.ref, g.url, g.path).CombinedOutput()
	if err != nil {
		log.Error(string(out))
	}
	return err
}
