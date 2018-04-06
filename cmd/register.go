package cmd

import (
	"errors"
	"io/ioutil"

	"github.com/mobingi/alm-agent/util"
	log "github.com/sirupsen/logrus"
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
		"echo /etc/sysctl.d/30-alm-agent.conf",
		"echo net.core.somaxconn = 65535 > /etc/sysctl.d/30-alm-agent.conf",
		"echo net.core.netdev_max_backlog = 20480 >> /etc/sysctl.d/30-alm-agent.conf",
		"echo net.ipv4.tcp_max_syn_backlog = 20480 >> /etc/sysctl.d/30-alm-agent.conf",
		"echo net.ipv4.tcp_tw_reuse = 1 >> /etc/sysctl.d/30-alm-agent.conf",
		"echo net.ipv4.ip_local_port_range = 10240 65535 >> /etc/sysctl.d/30-alm-agent.conf",
		"echo net.netfilter.nf_conntrack_max = 200000 >> /etc/sysctl.d/30-alm-agent.conf",
		"echo net.nf_conntrack_max = 200000 >> /etc/sysctl.d/30-alm-agent.conf",
		"sysctl -p -q /etc/sysctl.d/30-alm-agent.conf",
	}
	var out []byte

	provider := c.GlobalString("provider")
	switch provider {
	case "aws":
		cmdstrs = append(cmdstrs, "echo '* * * * * PATH=/sbin:/usr/bin:/bin /opt/mobingi/alm-agent/current/alm-agent -U ensure >> /var/log/alm-agent.log 2>&1' >> /tmp/crontab.alm-agent")
		cmdstrs = append(cmdstrs, "echo '* * * * * PATH=/sbin:/usr/bin:/bin /opt/mobingi/alm-agent/current/alm-agent-addon-aws >> /var/log/alm-agent/aws.log 2>&1' >> /tmp/crontab.alm-agent")
		err := putCheckConfig()
		if err == nil {
			cmdstrs = append(cmdstrs, "/sbin/chkconfig --add stop-alm-agent.sh")
			cmdstrs = append(cmdstrs, "/sbin/chkconfig stop-alm-agent.sh on")
			cmdstrs = append(cmdstrs, "/etc/init.d/stop-alm-agent.sh start")
		}
	case "alicloud":
		cmdstrs = append(cmdstrs, "echo '* * * * * PATH=/sbin:/usr/bin:/bin /opt/mobingi/alm-agent/current/alm-agent -P alicloud -U ensure >> /var/log/alm-agent.log 2>&1' >> /tmp/crontab.alm-agent")
	case "gcp":
		cmdstrs = append(cmdstrs, "echo '* * * * * PATH=/sbin:/usr/bin:/bin /opt/mobingi/alm-agent/current/alm-agent -P gcp -U ensure >> /var/log/alm-agent.log 2>&1' >> /tmp/crontab.alm-agent")
	case "k5":
		cmdstrs = append(cmdstrs, "echo '* * * * * PATH=/sbin:/usr/bin:/bin /opt/mobingi/alm-agent/current/alm-agent -P k5 -U ensure >> /var/log/alm-agent.log 2>&1' >> /tmp/crontab.alm-agent")
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

var chkconfigContent = `#!/bin/bash

# chkconfig:   2345 96 01
# description: stop-alm-agent

### BEGIN INIT INFO
# Provides: stop-alm-agent
# Required-Start: $local_fs $network $remote_fs
# Should-Start: $time
# Required-Stop: $local_fs $network $remote_fs
# Should-Stop:
# Default-Start: 2 3 4 5
# Default-Stop: 0 1 6
# Short-Description: stop-alm-agent
# Description: stop-alm-agent
### END INIT INFO

lock_file="/var/lock/subsys/stop-alm-agent"

start()
{
  touch ${lock_file}
}

stop()
{
  rm -rf ${lock_file}
  LANG=C date > /var/log/alm-agent-stop.last.log
  /opt/mobingi/alm-agent/current/alm-agent stop >> /var/log/alm-agent-stop.last.log 2>&1
}

case "$1" in
  start)
    start
  ;;
  stop)
    echo "invoke alm-agent stop ..."
    stop
  ;;
  *)
    echo "Usage: $0 {start|stop}"
  ;;
esac

exit 0
`

func putCheckConfig() error {
	err := ioutil.WriteFile("/etc/init.d/stop-alm-agent.sh", []byte(chkconfigContent), 00755)
	if err != nil {
		return err
	}
	return nil
}
