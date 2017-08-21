package versions

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-version"
	latest "github.com/tcnksm/go-latest"
)

// GoLatest uses check latest version
// Version infomations could be overwritten by LDFLAGS e.g) 'main.version=$(VERSION)'
type GoLatest struct {
	Version string `json:"version"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

var (
	// Version : majour.minor.epochtime
	Version = "0.1.1-dev"
	// Revision : Commit SHA1.
	Revision = "local-build"
	// URLBase : host version_info and binaries.
	URLBase = "https://download.labs.mobingi.com/alm-agent/"
	// Branch : current branch of build environment.
	Branch = "develop"
	// BinVer : path to latest symlink
	BinVer = "current"
)

var (
	basedir = "/opt/mobingi/alm-agent"
	keepNum = 5
)

// AutoUpdate checks latest version and replace to latest.
func AutoUpdate(v *GoLatest) {
	json := &latest.JSON{
		URL: v.URL,
	}

	res, err := latest.Check(json, v.Version)
	if err != nil {
		log.Warnf("AutoUpdate: Failed to check latest version. Autoupdate was skipped.")
		return
	}
	log.Debugf("AutoUpdate: Current Version is ", v.Version)
	if res.Outdated {
		log.Infof("%s is not latest, %s, upgrade to %s", v.Version, res.Meta.Message, res.Current)
		ensure(v, res.Current)
		removeOlders()
	} else {
		log.Debug("AutoUpdate: Using newest.")
	}

	return
}

func ensure(v *GoLatest, newVer string) {
	var err error

	os.MkdirAll(filepath.Join(basedir, "v"+newVer), 0700)
	tmpdir, _ := ioutil.TempDir("", "modaemon")
	defer os.RemoveAll(tmpdir)

	tmpPath := filepath.Join(tmpdir, "alm-agent.tgz")
	downloadLatest(tmpPath, newVer)

	file, _ := os.Open(tmpPath)
	defer file.Close()

	gzipReader, _ := gzip.NewReader(file)
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	log.Debugf("AutoUpdate: Extracting... %s", tmpPath)
	var header *tar.Header
	for {
		header, err = tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}

		buf := new(bytes.Buffer)
		if _, err = io.Copy(buf, tarReader); err != nil {
			log.Fatalln(err)
		}

		log.Debugf("AutoUpdate: Found %s", header.Name)
		if err = ioutil.WriteFile(basedir+"/"+header.Name, buf.Bytes(), 0755); err != nil {
			log.Fatal(err)
		}
	}

	symlinkPath := filepath.Join(basedir, BinVer)
	newVerPath := filepath.Join(basedir, "v"+newVer)
	if _, err := os.Lstat(symlinkPath); err == nil {
		os.Remove(symlinkPath)
	}
	os.Symlink(newVerPath, symlinkPath)

	log.Infof("AutoUpdate: Update Finished to v%s", newVer)
	return
}

func downloadLatest(tmpPath string, newVer string) {
	urlPath := strings.Join([]string{URLBase, Branch, "/v" + newVer + "/alm-agent.tgz"}, "")
	u, _ := url.Parse(urlPath)
	log.Infof("AutoUpdate: Trying to GET %s", u.String())
	req, _ := http.NewRequest("GET", u.String(), nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	file, _ := os.Create(tmpPath)
	io.Copy(file, res.Body)
	return
}

func removeOlders() {
	versionDirs, err := filepath.Glob(filepath.Join(basedir, "v*"))
	if err != nil {
		return
	}

	versionList := make([]*version.Version, len(versionDirs))
	for i, raw := range versionDirs {
		raws := strings.Split(raw, "/")
		v, _ := version.NewVersion(raws[len(raws)-1])
		versionList[i] = v
	}

	sort.Sort(version.Collection(versionList))
	if len(versionList) >= keepNum {
		log.Info("AutoUpdate: Cleans up olders.")
		for _, d := range versionList[:len(versionList)-keepNum] {
			dv := filepath.Join(basedir, "v"+d.String())
			log.Infof("AutoUpdate: Removing %s .", dv)
			os.RemoveAll(dv)
		}
	}

	return
}
