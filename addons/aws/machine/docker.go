package machine

import (
	"context"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/opts"
	"github.com/mobingilabs/go-modaemon/container"
	"github.com/mobingilabs/go-modaemon/server_config"
)

// ExecShutdownTaskOnAppContainers runs final tasks before shutdown instance.
func (m *Machine) ExecShutdownTaskOnAppContainers(s *serverConfig.Config) {
	d := container.NewDocker(s)
	log.Debugf("%#v", d)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := opts.NewFilterOpt()
	filter.Set(fmt.Sprintf("image=%s", s.Image))
	options := types.ContainerListOptions{
		Filters: filter.Value(),
	}
	lsres, err := d.client.ContainerList(context.Background(), options)
	if err != nil {
		log.Debugf("%#v", err)
	}
	log.Debugf("%#v", lsles)

	exc := types.ExecConfig{
		Cmd: []string{"/pre_shutdown.sh"},
	}

	res, err := d.client.ContainerExecCreate(ctx, "5543721c64e0127d2273f654601f4a511928344d6190ed0c1bab1b145dd9ef68", exc)
	if err != nil {
		log.Debugf("%#v", err)
	}

	esc := types.ExecStartCheck{
		Detach: true,
		Tty:    true,
	}

	err = client.ContainerExecStart(ctx, res.ID, esc)
	if err != nil {
		log.Warnf("%#v", err)
	}
	return
}
