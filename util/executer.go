package util

import (
	"os/exec"
	"strings"
)

// ExecuterInterface has Exec to wrap os commands.
type ExecuterInterface interface {
	Exec(cmd string, args ...string) ([]byte, error)
	ExecWithOpts(execopts *ExecOpts, cmd string, args ...string) ([]byte, error)
}

// ExecOpts is options for ExecWithOpts
type ExecOpts struct {
	Dir string
	Env []string
}

type osExecuter struct{}

// Executer is real Executer
var Executer ExecuterInterface

func init() {
	Executer = &osExecuter{}
}

func (o *osExecuter) Exec(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	return out, err
}

func (o *osExecuter) ExecWithOpts(opts *ExecOpts, command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = opts.Dir
	cmd.Env = opts.Env
	out, err := cmd.CombinedOutput()
	return out, err
}

// MockExecuter is fake Executer.
type MockExecuter struct{}

// MockBuffer collects commndline inputs.
var MockBuffer []string

// Exec returns command + args and put these to StdOut.
func (m *MockExecuter) Exec(command string, args ...string) ([]byte, error) {
	cl := command + " " + strings.Join(args, " ")
	MockBuffer = append(MockBuffer, cl)
	out := []byte(cl)
	return out, nil
}

// ExecWithOpts returns command + args and put these to StdOut.
func (m *MockExecuter) ExecWithOpts(opts *ExecOpts, command string, args ...string) ([]byte, error) {
	cl := command + " " + strings.Join(args, " ")
	MockBuffer = append(MockBuffer, cl)
	MockBuffer = append(MockBuffer, opts.Dir)
	MockBuffer = append(MockBuffer, strings.Join(opts.Env, ","))
	out := []byte(cl)
	return out, nil
}

// GetMockBuffer returns buffered commands
func GetMockBuffer() []string {
	return MockBuffer
}

// ClearMockBuffer discards buffer
func ClearMockBuffer() {
	MockBuffer = nil
	return
}
