package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	dproxy "github.com/koron/go-dproxy"
	"github.com/mobingi/alm-agent/api"
	"github.com/mobingi/alm-agent/bindata"
	"github.com/mobingi/alm-agent/code"
	"github.com/mobingi/alm-agent/config"
	"github.com/mobingi/alm-agent/container"
	"github.com/mobingi/alm-agent/login"
	"github.com/mobingi/alm-agent/server_config"
	"github.com/mobingi/alm-agent/util"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func tracerPath() string {
	selfPath, _ := os.Executable()
	d, _ := filepath.Split(selfPath)
	tracerPath := filepath.Join(d, "alm-logtracer")
	return tracerPath
}

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
	syscondata := bindata.Assets.Files["/_data/sys_containers.toml"].Data
	toml.Decode(string(syscondata), &syscons)

	addcons := &container.SystemContainers{}
	addcondata := bindata.Assets.Files["/_data/addon_containers.toml"].Data
	toml.Decode(string(addcondata), &addcons)

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
		api.SendContainerStatus("starting")
	}

	stsToken, err := api.GetStsToken()
	if err != nil {
		api.SendAgentStatus("error", err.Error())
		return cli.NewExitError(err, 1)
	}

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
		for _, y := range s.Users {
			login.EnsureUser(y.UserName, y.PublicKey)
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
			} else {
				log.Debugf("system container %s is up to date.", syscon.Name)
				continue
			}
		}

		sysContainer, _ = sc.StartSysContainer(&syscon)
		log.Debugf("%#v", sysContainer)
	}

	// Addon Containers
	for _, addon := range s.Addons {
		log.Debug("Step: NewAddonDockers")
		a := dproxy.New(addon)
		aName, err := a.M("name").String()
		if aName == "" || err != nil {
			log.Debugf("Failed?? to load addon container")
			continue
		}
		log.Debugf("%#v", addon)

		var addcon container.SystemContainer
		for _, con := range addcons.Container {
			if con.Name == aName {
				addcon = con
			} else {
				log.Errorf("Addon container %s not defined.", aName)
				continue
			}
		}

		ac, err := container.NewAddonDocker(conf, aName, addon, &addcon)
		if err != nil {
			log.Errorf("Failed to launch addon container")
			continue
		}

		log.Debugf("Step: ac.StartContainer")
		log.Debugf("%#v", ac)

		addonImageUpdated, err := ac.CheckImageUpdated()
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		addonContainer, err := ac.GetContainer(aName)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		log.Debugf("%#v", addonContainer)

		if addonContainer != nil {
			container.CompareToRestart(aName, ac, addonContainer, &addcon)
			if addonImageUpdated {
				ac.StopContainer(addonContainer)
				ac.RemoveContainer(addonContainer)
			} else if addcon.Restart {
				log.Debugf("addon container need to update by flag.", aName)
				ac.StopContainer(addonContainer)
				ac.RemoveContainer(addonContainer)
			} else if addonContainer.State == "exited" {
				ac.RemoveContainer(addonContainer)
			} else {
				log.Debugf("addon container %s is up to date.", aName)
				continue
			}
		}

		addonContainer, _ = ac.StartSysContainer(&addcon)
		log.Debugf("%#v", addonContainer)
	}

	// User Container
	log.Debug("Step: container.NewDocker")

	d, err := container.NewDocker(conf, s)
	if err != nil {
		api.SendAgentStatus("error", err.Error())
		return cli.NewExitError(err, 1)
	}
	log.Debugf("%#v", d)

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

	if !initialize {
		api.SendContainerStatus("updating")
	}

	codeDir := ""
	if s.GitRepo != "" {
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
	log.Debugf("CodeDir is %s", codeDir)

	if oldContainer != nil {
		d.MapPort(oldContainer) // For regenerating port map information
	}

	// standby exists?
	stc, _ := d.GetContainer("standby")
	if stc != nil {
		if stc.State == "exited" {
			d.RemoveContainer(stc)
		} else {
			d.StopContainer(stc)
			d.RemoveContainer(stc)
		}
	}

	log.Debug("Step: d.StartContainer")
	newContainer, err := d.StartContainer("standby", codeDir)
	if err != nil {
		api.SendAgentStatus("error", err.Error())
		return cli.NewExitError(err, 1)
	}
	log.Debugf("%#v", newContainer)

	log.Debug("Step: d.MapPort")
	d.UnmapPort()
	err = d.MapPort(newContainer)
	if err != nil {
		api.SendAgentStatus("error", err.Error())
		return cli.NewExitError(err, 1)
	}

	if oldContainer != nil {
		if oldContainer.State == "running" {
			d.StopContainer(oldContainer)
			d.RemoveContainer(oldContainer)
		} else if oldContainer.State == "exited" {
			d.RemoveContainer(oldContainer)
		}
	}

	d.RenameContainer(newContainer, "active")

	if util.FileExists(tracerPath()) {
		exec.Command(tracerPath(), newContainer.ID).Start()
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
