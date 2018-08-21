package machine

import (
	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/container"
	"github.com/mobingi/alm-agent/server_config"
)

// ExecShutdownTaskOnAppContainers runs final tasks before shutdown instance.
func (m *Machine) ExecShutdownTaskOnAppContainers(c *config.Config, s *serverConfig.Config) {
	d, _ := container.NewDocker(c, s)
	log.Debugf("%#v", d)

	conID, err := d.GetContainerIDbyImage(s.Image)
	if err != nil {
		log.Debugf("%#v", err)
	}
	log.Debugf("%#v", conID)

	res, err := d.CreateContainerExec(conID, []string{"/pre_shutdown.sh"})
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
