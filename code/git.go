package code

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mobingi/alm-agent/util"
	log "github.com/sirupsen/logrus"
)

// Git is wrapper of git command.
type Git struct {
	url  string
	path string
	ref  string
}

var isTag = regexp.MustCompile("^refs/tags/")

func (g *Git) getRemoteCommitHash() (string, error) {
	// git ls-remote origin master -q | cut -f 1
	opts := &util.ExecOpts{}
	opts.Dir = g.path

	out, err := execPipeline(
		g.path,
		[]string{"git", "ls-remote", "origin", g.ref},
		[]string{"cut", "-f", "1"},
	)
	if err != nil {
		log.Error(string(out))
		return "", err
	}

	remoteHash := strings.Trim(string(out), "\n")
	if remoteHash == "" {
		return "", fmt.Errorf("git ref %s not found", g.ref)
	}

	return remoteHash, nil
}

func (g *Git) fixOrigin() error {
	// check origin url and update remote by git remote set-url origin {{newUrl}}
	opts := &util.ExecOpts{}
	opts.Dir = g.path

	// same as git remote get-url, but older git does not support it
	out, err := execPipeline(
		g.path,
		[]string{"git", "remote", "-v"},
		[]string{"grep", "fetch"},
		[]string{"cut", "-f", "2"},
	)
	if err != nil {
		log.Error(string(out))
		return err
	}
	currentRemoteRaw := strings.Split(strings.Trim(string(out), "\n"), " ")
	currentRemote := currentRemoteRaw[0]

	if currentRemote != g.url {
		out, err = util.Executor.ExecWithOpts(opts, "git", "remote", "set-url", "origin", g.url)
		if err != nil {
			log.Error(string(out))
			return err
		}
	}

	return nil
}

func (g *Git) deepFetch(opts *util.ExecOpts) error {
	g.fixOrigin()

	// Fetch All
	// git fetch --prune
	// git tag -l | xargs git tag -d
	// git fetch --tags
	log.Debugf("Fetching remote %s", g.url)
	out, err := util.Executor.ExecWithOpts(opts, "git", "fetch", "--prune")
	if err != nil {
		log.Error(string(out))
		return err
	}

	_, err = execPipeline(
		g.path,
		[]string{"git", "tag", "-l"},
		[]string{"xargs", "git", "tag", "-d"},
	)
	if err != nil {
		log.Error(string(out))
		return err
	}

	out, err = util.Executor.ExecWithOpts(opts, "git", "fetch", "--tags")
	if err != nil {
		log.Error(string(out))
		return err
	}
	return nil
}

func (g *Git) get() error {
	opts := &util.ExecOpts{}
	opts.Env = []string{"GIT_SSH=" + filepath.Join(sshDir, gitSSHScriptName)}

	// Initial Clone as Bare repo
	if !util.FileExists(g.path) {
		log.Infof("Executing git clone %s %s", g.url, g.path)
		out, err := util.Executor.ExecWithOpts(opts, "git", "clone", "--bare", g.url, g.path)
		if err != nil {
			log.Error(string(out))
			return err
		}
	} else {
		// Update Cached-Copy
		opts.Dir = g.path
		g.deepFetch(opts)
	}

	return nil
}

func (g *Git) release(hash, releasePath string) error {
	if err := os.MkdirAll(releasePath, 0755); err != nil {
		return err
	}

	// git archive --format=tar P{hash}} | tar -C {{releasePath} -xf -
	out, err := execPipeline(
		g.path,
		[]string{"git", "archive", "--format=tar", hash},
		[]string{"tar", "-C", releasePath, "-xf", "-"},
	)

	if err != nil {
		log.Error(string(out))
		return err
	}

	log.Debug(out)
	return nil
}

func execPipeline(dir string, commands ...[]string) ([]byte, error) {
	cmds := make([]*exec.Cmd, len(commands))
	var err error

	for i, c := range commands {
		log.Debug(c)
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
