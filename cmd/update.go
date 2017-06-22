package cmd

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mobingilabs/go-modaemon/api"
	"github.com/mobingilabs/go-modaemon/code"
	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/container"
	"github.com/mobingilabs/go-modaemon/login"
	"github.com/mobingilabs/go-modaemon/util"
	"github.com/urfave/cli"
)

func Update(c *cli.Context) error {
	serverid, err := util.GetServerID()
	conf, err := config.LoadFromFile(c.String("config"))
	if err != nil {
		return err
	}

	apiClient, err := api.NewClient(conf)
	if err != nil {
		return err
	}

	stsToken, err := apiClient.GetStsToken()
	if err != nil {
		return err
	}

	apiClient.WriteTempToken(stsToken)

	s, err := apiClient.GetServerConfig(c.String("serverconfig"))
	if err != nil {
		return err
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

	d, err := container.NewDocker(conf, s)
	if err != nil {
		return err
	}

	apiClient.SendInstanceStatus(serverid, util.FetchContainerState())
	imageUpdated, err := d.CheckImageUpdated()
	if err != nil {
		return err
	}

	oldContainer, err := d.GetContainer("active")
	if err != nil {
		return err
	}

	if oldContainer == nil {
		return Start(c)
	}

	if !codeUpdated && !imageUpdated {
		return nil
	}

	apiClient.SendInstanceStatus(serverid, "updating")
	d.MapPort(oldContainer) // For regenerating port map information

	newContainer, err := d.StartContainer("standby", codeDir)
	if err != nil {
		return err
	}

	d.UnmapPort()
	d.MapPort(newContainer)

	d.StopContainer(oldContainer)
	d.RemoveContainer(oldContainer)

	d.RenameContainer(newContainer, "active")

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
				apiClient.SendInstanceStatus(serverid, s)
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
