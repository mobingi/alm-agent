package code

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/mobingi/alm-agent/util"
)

type Git struct {
	url  string
	path string
	ref  string
}

var isTag = regexp.MustCompile("^refs/tags/")

func (g *Git) checkUpdate() (bool, error) {
	out, err := execPipeline(
		g.path,
		[]string{"git", "remote", "-v"},
		[]string{"grep", "fetch"},
		[]string{"awk", "{print $2}"},
	)

	if err != nil {
		return false, err
	}

	url := strings.Trim(string(out), "\n")
	if url != g.url {
		return true, nil
	}

	opts := &util.ExecOpts{}
	opts.Env = []string{"GIT_SSH=" + filepath.Join(sshDir, gitSshScriptName)}
	opts.Dir = g.path

	out, err = util.Executor.ExecWithOpts(opts, "git", "fetch")
	if err != nil {
		log.Error(string(out))
	}

	if isTag.MatchString(g.ref) {
		out, err = util.Executor.ExecWithOpts(opts, "git", "diff", g.ref)
	} else {
		out, err = util.Executor.ExecWithOpts(opts, "git", "diff", fmt.Sprintf("origin/%s", g.ref))
	}

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
	opts := &util.ExecOpts{}
	opts.Env = []string{"GIT_SSH=" + filepath.Join(sshDir, gitSshScriptName)}

	if isTag.MatchString(g.ref) {
		log.Infof("Executing git clone %s %s", g.url, g.path)
		out, err := util.Executor.ExecWithOpts(opts, "git", "clone", g.url, g.path)
		if err != nil {
			log.Error(string(out))
		}

		log.Infof("Executing git checkout %s ", g.ref)
		opts.Dir = g.path
		out, err = util.Executor.ExecWithOpts(opts, "git", "checkout", g.ref)
		if err != nil {
			log.Error(string(out))
		}
		return err
	} else {
		log.Infof("Executing git clone -b %s %s %s", g.ref, g.url, g.path)
		out, err := util.Executor.ExecWithOpts(opts, "git", "clone", "-b", g.ref, g.url, g.path)
		if err != nil {
			log.Error(string(out))
		}
		return err
	}
}

func execPipeline(dir string, commands ...[]string) ([]byte, error) {
	cmds := make([]*exec.Cmd, len(commands))
	var err error

	for i, c := range commands {
		cmds[i] = exec.Command(c[0], c[1:]...)
		if dir != "" {
			cmds[i].Dir = dir
		}
		if i > 0 {
			if cmds[i].Stdin, err = cmds[i-1].StdoutPipe(); err != nil {
				return nil, err
			}
		}
		cmds[i].Stderr = os.Stderr
	}
	var out bytes.Buffer
	cmds[len(cmds)-1].Stdout = &out
	for _, c := range cmds {
		if err = c.Start(); err != nil {
			return nil, err
		}
	}
	for _, c := range cmds {
		if err = c.Wait(); err != nil {
			return nil, err
		}
	}
	return out.Bytes(), nil
}
