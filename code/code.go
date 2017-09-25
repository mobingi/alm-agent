package code

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingi/alm-agent/server_config"
	"github.com/mobingi/alm-agent/util"
)

var (
	knownHosts       = "/etc/ssh/ssh_known_hosts"
	sshDir           = "/opt/mobingi/etc/ssh"
	sshKeyName       = "id_alm_agent"
	gitSshScriptName = "git_ssh.sh"
)

// ref. `man git-clone`
// The following syntaxes may be used.
// - ssh://[user@]host.xz[:port]/path/to/repo.git/
// - git://host.xz[:port]/path/to/repo.git/
// - http[s]://host.xz[:port]/path/to/repo.git/
// - ftp[s]://host.xz[:port]/path/to/repo.git/
// - [user@]host.xz:path/to/repo.git/
var (
	hasSchemeSyntax = regexp.MustCompile("^[^:]+://")
	scpLikeSyntax   = regexp.MustCompile("^([^@]+@)?([^:]+):/?(.+)$")
)

type Code struct {
	URL  string
	Ref  string
	Path string
	Key  string
}

type Dirs []os.FileInfo

func (d Dirs) Len() int {
	return len(d)
}

func (d Dirs) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
func (d Dirs) Less(i, j int) bool {
	return d[j].ModTime().Unix() < d[i].ModTime().Unix()
}

func New(s *serverConfig.Config) *Code {
	ref := s.GitReference
	if ref == "" {
		ref = "master"
	}
	return &Code{
		URL: s.GitRepo,
		Ref: ref,
		Key: s.GitPrivateKey,
	}
}

func (c *Code) CheckUpdate() (bool, error) {
	base := "/srv/code"
	if !util.FileExists(base) {
		return true, nil
	}

	dirs, err := ioutil.ReadDir(base)

	if err != nil {
		return false, err
	}

	if len(dirs) == 0 {
		return true, nil
	}

	sort.Sort(Dirs(dirs))

	if len(dirs) > 10 {
		for _, dir := range dirs[10:] {
			err := os.RemoveAll(filepath.Join(base, dir.Name()))
			if err != nil {
				return false, err
			}
		}
	}

	c.Path = filepath.Join(base, dirs[0].Name())
	g := &Git{
		url:  c.URL,
		path: c.Path,
		ref:  c.Ref,
	}

	return g.checkUpdate()
}

func (c *Code) Get() (string, error) {
	baseDir := filepath.Join("/srv", "code")
	if !util.FileExists(baseDir) {
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return "", err
		}
	}

	t := time.Now().Format("20060102150405")
	g := &Git{
		url:  c.URL,
		path: filepath.Join(baseDir, t),
		ref:  c.Ref,
	}
	err := g.get()
	return g.path, err
}

func (c *Code) PrivateRepo() error {
	err := createIdentityFile(c.Key)
	if err != nil {
		return err
	}

	url, err := parseURL(c.URL)
	if err != nil {
		return err
	}

	if url.Scheme == "git" && url.Host == "github.com" {
		c.URL = convertGithubGitURLToSSH(url)
		log.Debugf("Converted URL is %s", c.URL)
	}

	err = checkKnownHosts(url)
	if err != nil {
		return err
	}

	err = writeGitSshScript()
	if err != nil {
		return err
	}

	return nil
}

func createIdentityFile(key string) error {
	log.Debug("Step: createIdentityFile")
	if !util.FileExists(sshDir) {
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			return err
		}
	}

	sshKey := filepath.Join(sshDir, sshKeyName)

	log.Debugf("Create IdentityFile %s", sshKey)
	err := ioutil.WriteFile(sshKey, []byte(key), 0600)
	if err != nil {
		return err
	}
	return nil
}

func parseURL(rawURL string) (*url.URL, error) {
	log.Debug("Step: parseURL")
	if !hasSchemeSyntax.MatchString(rawURL) && scpLikeSyntax.MatchString(rawURL) {
		matched := scpLikeSyntax.FindStringSubmatch(rawURL)
		user := matched[1]
		host := matched[2]
		path := matched[3]
		rawURL = fmt.Sprintf("ssh://%s%s/%s", user, host, path)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return u, err
	}

	return u, nil
}

func convertGithubGitURLToSSH(url *url.URL) string {
	return fmt.Sprintf("ssh://git@github.com%s", url.Path)
}

func checkKnownHosts(url *url.URL) error {
	log.Debug("Step: checkKnownHosts")
	out, err := util.Executer.Exec("ssh-keygen", "-F", url.Host, "-f", knownHosts)
	if string(out) == "" && err != nil {
		out, err := util.Executer.Exec("ssh-keyscan", url.Host)
		if err != nil {
			return err
		}
		if string(out) == "" {
			return fmt.Errorf("%s's ssh public key is empty", url.Host)
		}

		log.Debugf("Add %s's public key to %s", url.Host, knownHosts)

		file, err := os.OpenFile(knownHosts, os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		io.WriteString(file, string(out))
		return nil
	}
	return nil
}

func writeGitSshScript() error {
	log.Debug("Step: writeGitSshScript")
	c := `#!/bin/sh
exec ssh -i %s "$@"
`
	s := fmt.Sprintf(c, filepath.Join(sshDir, sshKeyName))
	err := ioutil.WriteFile(filepath.Join(sshDir, gitSshScriptName), []byte(s), 0700)
	if err != nil {
		return err
	}

	return nil
}
