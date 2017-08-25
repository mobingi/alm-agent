package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mobingi/alm-agent/util"
	"github.com/urfave/cli"
)

// Register alm-agent register
func Register(c *cli.Context) error {
	// TODO: refactor it !!
	var cmdstr string
	var out []byte

	// mkdirs
	log.Warn("mkdirs")
	cmdstr = "mkdir -p /var/log/alm-agent/containerlogs /var/log/alm-agent/container"
	out, _ = util.Executer.Exec("sh", "-c", cmdstr)
	log.Debug(string(out))

	// known_hosts
	log.Warn("set known_hosts")
	cmdstr = "ssh-keyscan -t rsa -H github.com | tee /etc/ssh/ssh_known_hosts"
	out, _ = util.Executer.Exec("sh", "-c", cmdstr)
	log.Debug(string(out))

	cmdstr = "ssh-keyscan -t dsa -H github.com | tee -a /etc/ssh/ssh_known_hosts"
	out, _ = util.Executer.Exec("sh", "-c", cmdstr)
	log.Debug(string(out))

	cmdstr = "ssh-keyscan -t rsa -H bitbucket.org | tee -a /etc/ssh/ssh_known_hosts"
	out, _ = util.Executer.Exec("sh", "-c", cmdstr)
	log.Debug(string(out))

	cmdstr = "ssh-keyscan -t dsa -H bitbucket.org | tee -a /etc/ssh/ssh_known_hosts"
	out, _ = util.Executer.Exec("sh", "-c", cmdstr)
	log.Debug(string(out))

	cmdstr = "ssh-keyscan -t rsa -H gitlab.com | tee -a /etc/ssh/ssh_known_hosts"
	out, _ = util.Executer.Exec("sh", "-c", cmdstr)
	log.Debug(string(out))

	cmdstr = "ssh-keyscan -t dsa -H gitlab.com | tee -a /etc/ssh/ssh_known_hosts"
	out, _ = util.Executer.Exec("sh", "-c", cmdstr)
	log.Debug(string(out))

	// crontab
	log.Warn("crontab entry")
	cmdstr = "echo '* * * * * /opt/mobingi/alm-agent/current/alm-agent -U ensure >> /var/log/alm-agent.log 2&>1' > /tmp/crontab.alm-agent"
	out, _ = util.Executer.Exec("sh", "-c", cmdstr)
	log.Debug(string(out))
	cmdstr = "crontab /tmp/crontab.alm-agent"
	out, _ = util.Executer.Exec("sh", "-c", cmdstr)
	log.Debug(string(out))
	cmdstr = "rm -f /tmp/crontab.alm-agent"
	out, _ = util.Executer.Exec("sh", "-c", cmdstr)
	log.Debug(string(out))

	err := Ensure(c)
	return err
}
