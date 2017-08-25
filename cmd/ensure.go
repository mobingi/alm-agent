package cmd

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingi/alm-agent/api"
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

	serverid, err := util.GetServerID(c.GlobalString("provider"))
	if err != nil {
		return err
	}

	// detect where come from to merge Start and Update
	initialize = (c.Command.Name == "register")

	log.Debug("Step: config.LoadFromFile")

	conf, err := config.LoadFromFile(c.String("config"))
	if err != nil {
		return err
	}

	api.SetConfig(conf)
	err = api.GetAccessToken()
	if err != nil {
		return err
	}

	if initialize {
		api.SendInstanceStatus(serverid, "starting")
	}

	if initialize {
		Start(c)
	} else {
		Update(c)
	}
	return nil

	stsToken, err := api.GetStsToken()
	if err != nil {
		return err
	}

	api.WriteTempToken(stsToken)

	s, err := api.GetServerConfig(c.String("serverconfig"))
	if err != nil {
		return err
	}

	ld, err := container.NewSysDocker(conf, serverid)
	if err != nil {
		return err
	}

	logImageUpdated, err := ld.CheckImageUpdated()
	if err != nil {
		return err
	}

	logContainer, err := ld.GetContainer("alm-awslogs")
	if err != nil {
		return err
	}

	d, err := container.NewDocker(conf, s)
	if err != nil {
		return err
	}

	oldContainer, err := d.GetContainer("active")
	if err != nil {
		return err
	}

	if logContainer == nil && oldContainer == nil {
		return Start(c)
	}

	if logImageUpdated {
		ld.StopContainer(logContainer)
		ld.RemoveContainer(logContainer)
		_, err := ld.StartContainer("alm-awslogs", "", false)
		if err != nil {
			return err
		}
	}

	update, err := serverConfig.NeedsUpdate(s)
	if err != nil {
		return err
	}
	if !update {
		return nil
	}

	for x, y := range s.Users {
		login.EnsureUser(x, y.PublicKey)
	}

	codeDir := ""
	codeUpdated := false
	if s.Code != "" {
		code := code.New(s)
		if code.Key != "" {
			err = code.PrivateRepo()
			if err != nil {
				return err
			}
		}

		codeUpdated, err = code.CheckUpdate()
		if err != nil {
			return err
		}

		if codeUpdated {
			codeDir, err = code.Get()
			if err != nil {
				return err
			}
		} else {
			codeDir = code.Path
		}
	}

	api.SendInstanceStatus(serverid, "updating")
	d.MapPort(oldContainer) // For regenerating port map information

	newContainer, err := d.StartContainer("standby", codeDir, true)
	if err != nil {
		return err
	}

	d.UnmapPort()
	d.MapPort(newContainer)

	d.StopContainer(oldContainer)
	d.RemoveContainer(oldContainer)

	d.RenameContainer(newContainer, "active")

	if err := serverConfig.WriteUpdated(s); err != nil {
		return err
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
				return
			case s := <-state:
				if s != "" {
					api.SendInstanceStatus(serverid, s)
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
	return nil
}
