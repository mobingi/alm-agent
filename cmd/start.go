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

// Start alm-agent start
func Start(c *cli.Context) error {
	log.Debug("Step: config.LoadFromFile")

	serverid, err := util.GetServerID(c.GlobalString("provider"))
	if err != nil {
		return err
	}

	conf, err := config.LoadFromFile(c.String("config"))
	if err != nil {
		return err
	}
	log.Debugf("%#v", conf)
	api.SetConfig(conf)
	err = api.GetAccessToken()
	if err != nil {
		return err
	}

	api.SendInstanceStatus(serverid, "starting")

	stsToken, err := api.GetStsToken()
	if err != nil {
		api.SendInstanceStatus(serverid, "error")
		return err
	}

	api.WriteTempToken(stsToken)

	log.Debug("Step: apiClient.GetServerConfig")
	log.Debugf("Flag: %#v", c.String("serverconfig"))
	s, err := api.GetServerConfig(c.String("serverconfig"))
	if err != nil {
		api.SendInstanceStatus(serverid, "error")
		return err
	}
	log.Debugf("%#v", s)

	for x, y := range s.Users {
		login.EnsureUser(x, y.PublicKey)
	}

	codeDir := ""
	if s.Code != "" {
		code := code.New(s)
		if code.Key != "" {
			log.Debug("Step: code.PrivateRepo")
			err = code.PrivateRepo()
			if err != nil {
				return err
			}
		}

		codeDir, err = code.Get()
		if err != nil {
			return err
		}
	}

	log.Debug("Step: NewSysDocker")
	ld, err := container.NewSysDocker(conf, serverid)
	if err != nil {
		return err
	}
	log.Debugf("%#v", ld)

	log.Debug("Step: ld.StartContainer")
	logContainer, err := ld.StartContainer("mo-awslogs", "", false)
	if err != nil {
		return err
	}
	log.Debugf("%#v", logContainer)

	log.Debug("Step: container.NewDocker")
	d, err := container.NewDocker(conf, s)
	if err != nil {
		api.SendInstanceStatus(serverid, "error")
		return err
	}
	log.Debugf("%#v", d)

	log.Debug("Step: d.StartContainer")
	newContainer, err := d.StartContainer("active", codeDir, true)
	if err != nil {
		api.SendInstanceStatus(serverid, "error")
		return err
	}
	log.Debugf("%#v", newContainer)

	log.Debug("Step: d.MapPort")
	err = d.MapPort(newContainer)
	if err != nil {
		api.SendInstanceStatus(serverid, "error")
		return err
	}

	log.Debug("Step: serverConfig.WriteUpdated")
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
				log.Error("Container start processing timed out.")
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
