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
	"strings"
	"time"

	"github.com/mobingi/alm-agent/server_config"
	"github.com/mobingi/alm-agent/util"
	log "github.com/sirupsen/logrus"
)

var (
	knownHosts       = "/etc/ssh/ssh_known_hosts"
	sshDir           = "/opt/mobingi/etc/ssh"
	sshKeyName       = "id_alm_agent"
	gitSSHScriptName = "git_ssh.sh"
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
var baseDir = "/srv/code"
var cacheDir = filepath.Join(baseDir, "cached-copy")
var releaseDir = filepath.Join(baseDir, "releases")

// Code is application repository
type Code struct {
	URL           string
	Ref           string
	Path          string
	Key           string
	LatestRelease string
	Updated       bool
}

// Release is exported code
type Release struct {
	Hash string
	Path string
}

func newRelease(hash string) *Release {
	t := time.Now().Format("20060102150405")
	releasePath := filepath.Join(releaseDir, t)
	return &Release{
		Hash: hash,
		Path: releasePath,
	}
}

func (c *Code) loadCurrentRelease() *Release {
	current := filepath.Join(baseDir, "current")
	r := &Release{}
	if util.FileExists(current) {
		releaseInfoRaw, _ := ioutil.ReadFile(current)
		releaseInfo := strings.Split(strings.Trim(string(releaseInfoRaw), "\n"), ",")

		// avoid format error
		if len(releaseInfo) == 2 {
			r.Hash = releaseInfo[0]
			r.Path = releaseInfo[1]
		}
	}
	return r
}

func (r *Release) putAsCurrent() {
	releaseInfo := strings.Join([]string{r.Hash, r.Path}, ",")
	ioutil.WriteFile(filepath.Join(baseDir, "current"), []byte(releaseInfo), 0644)
}

// ReleaseDirs are directories under /srv/code/releases
type ReleaseDirs []os.FileInfo

// Len to use ReleaseDirs as sort Interface.
// do not remove
func (d ReleaseDirs) Len() int {
	return len(d)
}

// Swap to use ReleaseDirs as sort Interface.
// do not remove
func (d ReleaseDirs) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Less to use ReleaseDirs as sort Interface.
// do not remove
func (d ReleaseDirs) Less(i, j int) bool {
	return d[j].ModTime().Unix() < d[i].ModTime().Unix()
}

// New creates New Code obj
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

// CheckUpdate checks code and cleans up old releases
func (c *Code) cleanupReleases() error {
	if !util.FileExists(releaseDir) {
		return nil
	}

	dirs, err := ioutil.ReadDir(releaseDir)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		return nil
	}

	fmt.Println(dirs)
	sort.Sort(ReleaseDirs(dirs))

	if len(dirs) > 5 {
		for _, dir := range dirs[5:] {
			delDirPath := filepath.Join(releaseDir, dir.Name())
			log.Infof("Cleaning up old release... %s", delDirPath)
			err := os.RemoveAll(delDirPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Get creates releases
// returns code dirpash to mount by container.
func (c *Code) Get() (string, error) {
	log.Debug("Code: Get")
	if !util.FileExists(releaseDir) {
		if err := os.MkdirAll(releaseDir, 0755); err != nil {
			return "", err
		}
	}

	g := &Git{
		url:  c.URL,
		path: cacheDir,
		ref:  c.Ref,
	}
	err := g.get()
	revCurrent := c.loadCurrentRelease().Hash
	revRemote, err := g.getRemoteCommitHash()
	if err != nil {
		return "", err
	}

	log.Debugf("Current: %s , Remote: %s", revCurrent, revRemote)
	if revCurrent != revRemote {
		re := newRelease(revRemote)
		if err := g.release(revRemote, re.Path); err != nil {
			return "", err
		}
		re.putAsCurrent()
		c.Updated = true
		c.cleanupReleases()
		return re.Path, nil
	}

	re := c.loadCurrentRelease()
	return re.Path, nil
}

// PrivateRepo sets up remote credential
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

	err = writeGitSSHScript()
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
	out, err := util.Executor.Exec("ssh-keygen", "-F", url.Host, "-f", knownHosts)
	if string(out) == "" && err != nil {
		out, err := util.Executor.Exec("ssh-keyscan", url.Host)
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

func writeGitSSHScript() error {
	log.Debug("Step: writeGitSshScript")
	c := `#!/bin/sh
exec ssh -i %s "$@"
`
	s := fmt.Sprintf(c, filepath.Join(sshDir, sshKeyName))
	err := ioutil.WriteFile(filepath.Join(sshDir, gitSSHScriptName), []byte(s), 0700)
	if err != nil {
		return err
	}

	return nil
}
