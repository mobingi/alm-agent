package cmd

import (
	"errors"

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
		"ssh-keyscan -t dsa -H github.com | tee -a /etc/ssh/ssh_known_hosts",
		"ssh-keyscan -t rsa -H bitbucket.org | tee -a /etc/ssh/ssh_known_hosts",
		"ssh-keyscan -t dsa -H bitbucket.org | tee -a /etc/ssh/ssh_known_hosts",
		"ssh-keyscan -t rsa -H gitlab.com | tee -a /etc/ssh/ssh_known_hosts",
		"ssh-keyscan -t dsa -H gitlab.com | tee -a /etc/ssh/ssh_known_hosts",
		"crontab -l | grep -v 'current/alm-agent' > /tmp/crontab.alm-agent",
	}
	var out []byte

	provider := c.GlobalString("provider")
	switch provider {
	case "aws":
		cmdstrs = append(cmdstrs, "echo '* * * * * PATH=/sbin:/usr/bin:/bin /opt/mobingi/alm-agent/current/alm-agent -U ensure >> /var/log/alm-agent.log 2>&1' >> /tmp/crontab.alm-agent")
		cmdstrs = append(cmdstrs, "echo '* * * * * PATH=/sbin:/usr/bin:/bin /opt/mobingi/alm-agent/current/alm-agent-addon-aws >> /var/log/alm-agent/aws.log 2>&1' >> /tmp/crontab.alm-agent")
	case "alicloud":
		cmdstrs = append(cmdstrs, "echo '* * * * * PATH=/sbin:/usr/bin:/bin /opt/mobingi/alm-agent/current/alm-agent -P alicloud -U ensure >> /var/log/alm-agent.log 2>&1' >> /tmp/crontab.alm-agent")
	case "localtest":
		return nil
	default:
		return cli.NewExitError(errors.New("Provider `"+provider+"` is not supported."), 1)
	}
	cmdstrs = append(cmdstrs, "crontab /tmp/crontab.alm-agent")
	cmdstrs = append(cmdstrs, "rm -f /tmp/crontab.alm-agent")

	for _, cmdstr := range cmdstrs {
		out, _ = util.Executor.Exec("sh", "-c", cmdstr)
		log.Debug(string(out))
	}

	err := Ensure(c)
	if err != nil {
		return err
	}
	return nil
}
