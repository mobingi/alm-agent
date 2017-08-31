package container

// VolHandle builds []string to volume mount.
type VolHandle func() []string

// VolFuncs contains funcs of VolHandle
type VolFuncs struct{}

var volFuncs = &VolFuncs{}

var handleVolMap = map[string]VolHandle{
	"logs_vol": volFuncs.LogsVol,
}

// LogsVol resolves logs_vol
func (v *VolFuncs) LogsVol() []string {
	vols := []string{
		"/root/.aws/awslogs_creds.conf:/etc/awslogs/awscli.conf",
		"/var/log:/var/log",
		containerLogsLocation + ":/var/container",
		"/opt/awslogs:/var/lib/awslogs",
	}
	return vols
}
