package util

import "os/exec"

// ExecuterInterface has Exec to wrap os commands.
type ExecuterInterface interface {
	Exec(string, ...string) ([]byte, error)
}

type osExecuter struct{}

// OSExecuter is real Executer
var Executer ExecuterInterface

func init() {
	Executer = &osExecuter{}
}

func (o *osExecuter) Exec(command string, args ...string) ([]byte, error) {
	out, err := exec.Command(command, args...).CombinedOutput()
	return out, err
}
