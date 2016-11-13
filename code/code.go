package code

import (
	"path"
	"time"

	"github.com/mobingilabs/go-modaemon/server_config"
)

func Get(s serverConfig.Config) (string, error) {
	t := time.Now().Format("20060102150405")
	g := &Git{
		url:  s.Code,
		path: path.Join("/srv", "code", t),
		ref:  s.GitReference,
	}
	err := g.get()
	return g.path, err
}
