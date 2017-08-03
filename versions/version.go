package versions

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/tcnksm/go-latest"
)

// GoLatest uses check latest version
// Version infomations could be overwritten by LDFLAGS e.g) 'main.version=$(VERSION)'
type GoLatest struct {
	Version string `json:"version"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

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
		fmt.Printf("%s is not latest, %s, upgrade to %s\n", v.Version, res.Meta.Message, res.Current)
	}
	return
}
