package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mobingi/alm-agent/util"
	"github.com/urfave/cli"
)

// Register alm-agent register
func Register(c *cli.Context) error {
	// TODO: refactor it !!
	var cmdstrs = []string{
		"mkdir -p /var/log/alm-agent/containerlogs /var/log/alm-agent/container",
		"ssh-keyscan -t rsa -H github.com | tee /etc/ssh/ssh_known_hosts",
		"ssh-keyscan -t dsa -H github.com | tee /etc/ssh/ssh_known_hosts",
		"ssh-keyscan -t rsa -H bitbucket.org | tee -a /etc/ssh/ssh_known_hosts",
		"ssh-keyscan -t dsa -H bitbucket.org | tee -a /etc/ssh/ssh_known_hosts",
		"ssh-keyscan -t rsa -H gitlab.com | tee -a /etc/ssh/ssh_known_hosts",
		"ssh-keyscan -t dsa -H gitlab.com | tee -a /etc/ssh/ssh_known_hosts",
		"echo '* * * * * /opt/mobingi/alm-agent/current/alm-agent -U ensure >> /var/log/alm-agent.log 2&>1' > /tmp/crontab.alm-agent",
		"crontab /tmp/crontab.alm-agent",
		"rm -f /tmp/crontab.alm-agent",
	}
	var out []byte

	for _, cmdstr := range cmdstrs {
		out, _ = util.Executer.Exec("sh", "-c", cmdstr)
		log.Debug(string(out))
	}

	err := Ensure(c)
	return err
}
