package code

import "os/exec"

type Git struct {
	url  string
	path string
	ref  string
}

func (g *Git) get() error {
	return exec.Command("git", "clone", "-b", g.ref, g.url, g.path).Run()
}
