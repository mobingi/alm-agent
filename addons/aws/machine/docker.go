package machine

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/mobingilabs/go-modaemon/container"
	"github.com/mobingilabs/go-modaemon/server_config"
)

// ExecShutdownTaskOnAppContainers runs final tasks before shutdown instance.
func (m *Machine) ExecShutdownTaskOnAppContainers(s *serverConfig.Config) {
	d, _ := container.NewDocker(s)
	log.Debugf("%#v", d)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conID := d.GetContainerIDbyImage(s.Image)
	if err != nil {
		log.Debugf("%#v", err)
	}
	log.Debugf("%#v", lsles)

	exc := types.ExecConfig{
		Cmd: []string{"/pre_shutdown.sh"},
	}

	res, err := d.CreateContainerExec(conID, exc)
	if err != nil {
		log.Debugf("%#v", err)
	}

	esc := types.ExecStartCheck{
		Detach: true,
		Tty:    true,
	}

	err = d.StartContainerExec(res.ID, esc)
	if err != nil {
		log.Warnf("%#v", err)
	}
	return
}
