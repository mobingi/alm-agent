package cmd

import (
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/mobingi/alm-agent/api"
	"github.com/mobingi/alm-agent/bindata"
	"github.com/mobingi/alm-agent/code"
	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/container"
	"github.com/mobingi/alm-agent/login"
	"github.com/mobingi/alm-agent/server_config"
	"github.com/mobingi/alm-agent/util"
	"github.com/urfave/cli"
)

// Ensure start or replace container with newest config
func Ensure(c *cli.Context) error {
	var initialize bool
	var err error

	initialize = (c.Command.Name == "register")
	util.AgentID()

	err = util.GetServerID(c.GlobalString("provider"))
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	conf, err := config.LoadFromFile(c.String("config"))

	syscons := &container.SystemContainers{}
	syscondata, _ := bindata.Asset("_data/sys_containers.toml")
	toml.Decode(string(syscondata), &syscons)

	if err != nil {
		return cli.NewExitError(err, 1)
	}
	log.Debugf("%#v", conf)
	api.SetConfig(conf)
	err = api.GetAccessToken()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	if initialize {
		api.SendAgentStatus("starting", "")
	}

	stsToken, err := api.GetStsToken()
	if err != nil {
		api.SendAgentStatus("error", err.Error())
		return cli.NewExitError(err, 1)
	}

	api.WriteTempToken(stsToken)

	log.Debug("Step: api.GetServerConfig")
	log.Debugf("Flag: %#v", c.String("serverconfig"))
	s, err := api.GetServerConfig(c.String("serverconfig"))
	if err != nil {
		api.SendAgentStatus("error", err.Error())
		return cli.NewExitError(err, 1)
	}
	log.Debugf("%#v", s)

	update, err := serverConfig.NeedsUpdate(s)
	if update {
		for x, y := range s.Users {
			login.EnsureUser(x, y.PublicKey)
		}
	}

	// System Containers
	log.Debug("Step: NewSysDockers")
	for _, syscon := range syscons.Container {
		log.Debugf("%#v", syscon)
		sc, err := container.NewSysDocker(conf, &syscon)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		log.Debugf("Step: sc.StartContainer")
		log.Debugf("%#v", sc)

		sysImageUpdated, err := sc.CheckImageUpdated()
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		sysContainer, err := sc.GetContainer(syscon.Name)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		log.Debugf("%#v", sysContainer)

		if sysContainer != nil {
			if sysImageUpdated {
				sc.StopContainer(sysContainer)
				sc.RemoveContainer(sysContainer)
			} else if sysContainer.State == "exited" {
				sc.RemoveContainer(sysContainer)
			}
		}

		sysContainer, _ = sc.StartSysContainer(&syscon)
		log.Debugf("%#v", sysContainer)
	}

	if initialize {
		// All of old Start command
		codeDir := ""
		if s.Code != "" {
			code := code.New(s)
			if code.Key != "" {
				log.Debug("Step: code.PrivateRepo")
				err = code.PrivateRepo()
				if err != nil {
					return cli.NewExitError(err, 1)
				}
			}

			codeDir, err = code.Get()
			if err != nil {
				return cli.NewExitError(err, 1)
			}
		}

		// User Container
		log.Debug("Step: container.NewDocker")
		api.SendContainerStatus("starting")

		d, err := container.NewDocker(conf, s)
		if err != nil {
			api.SendAgentStatus("error", err.Error())
			return cli.NewExitError(err, 1)
		}
		log.Debugf("%#v", d)

		log.Debug("Step: d.StartContainer")
		newContainer, err := d.StartContainer("active", codeDir)
		if err != nil {
			api.SendAgentStatus("error", err.Error())
			return cli.NewExitError(err, 1)
		}
		log.Debugf("%#v", newContainer)

		log.Debug("Step: d.MapPort")
		err = d.MapPort(newContainer)
		if err != nil {
			api.SendAgentStatus("error", err.Error())
			return cli.NewExitError(err, 1)
		}
	} else {
		// All of old Update commdnad
		d, err := container.NewDocker(conf, s)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		oldContainer, err := d.GetContainer("active")
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		if oldContainer == nil || oldContainer.State == "exited" {
			update = true
		}

		if !update {
			return nil
		}

		codeDir := ""
		codeUpdated := false
		if s.Code != "" {
			code := code.New(s)
			if code.Key != "" {
				err = code.PrivateRepo()
				if err != nil {
					return cli.NewExitError(err, 1)
				}
			}

			codeUpdated, err = code.CheckUpdate()
			if err != nil {
				return cli.NewExitError(err, 1)
			}

			if codeUpdated {
				codeDir, err = code.Get()
				if err != nil {
					return cli.NewExitError(err, 1)
				}
			} else {
				codeDir = code.Path
			}
		}

		api.SendContainerStatus("updating")
		if oldContainer != nil {
			d.MapPort(oldContainer) // For regenerating port map information
		}

		// standby exists?
		stc := d.GetContainer("standby")
		if stc != nil {
			if stc.State == "exited" {
				d.RemoveContainer(stc)
			} else {
				d.StopContainer(stc)
				d.RemoveContainer(stc)
			}
		}

		newContainer, err := d.StartContainer("standby", codeDir)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		d.UnmapPort()
		d.MapPort(newContainer)

		if oldContainer != nil && oldContainer.State == "running" {
			d.StopContainer(oldContainer)
			d.RemoveContainer(oldContainer)
		} else if oldContainer.State == "exited" {
			d.RemoveContainer(oldContainer)
		}

		d.RenameContainer(newContainer, "active")
	}

	log.Debug("Step: serverConfig.WriteUpdated")
	if err := serverConfig.WriteUpdated(s); err != nil {
		return cli.NewExitError(err, 1)
	}

	var wg sync.WaitGroup
	timer := time.NewTimer(180 * time.Second)
	state := make(chan string)
	done := make(chan bool)
	cancel := make(chan bool)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-cancel:
				log.Error("Container update processing timed out.")
				api.SendContainerStatus("unknown")
				api.SendAgentStatus("error", "Container update processing timed out.")
				return
			case s := <-state:
				if s != "" {
					api.SendContainerStatus(s)
				}
				if s == "complete" {
					done <- true
					return
				}
			}
		}
	}()

LOOP:
	for {
		select {
		case <-timer.C:
			cancel <- true
			break LOOP
		case <-done:
			break LOOP
		case state <- util.FetchContainerState():
			time.Sleep(2 * time.Second)
		}
	}

	wg.Wait()
	api.SendAgentStatus("uptodate", "")
	return nil
}
