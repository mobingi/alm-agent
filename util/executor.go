package util

import (
	"os"
	"os/exec"
	"strings"
)

// ExecutorInterface has Exec to wrap os commands.
type ExecutorInterface interface {
	Exec(cmd string, args ...string) ([]byte, error)
	ExecWithOpts(execopts *ExecOpts, cmd string, args ...string) ([]byte, error)
}

// ExecOpts is options for ExecWithOpts
type ExecOpts struct {
	Dir string
	Env []string
}

type osExecutor struct{}

// Executor is real Executor
var Executor ExecutorInterface

func init() {
	Executor = &osExecutor{}
}

func (o *osExecutor) Exec(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	return out, err
}

func (o *osExecutor) ExecWithOpts(opts *ExecOpts, command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = opts.Dir
	cmd.Env = append(os.Environ(), opts.Env...)
	out, err := cmd.CombinedOutput()
	return out, err
}

// MockExecutor is fake Executor.
type MockExecutor struct{}

// MockBuffer collects commndline inputs.
var MockBuffer []string

// Exec returns command + args and put these to StdOut.
func (m *MockExecutor) Exec(command string, args ...string) ([]byte, error) {
	cl := command + " " + strings.Join(args, " ")
	MockBuffer = append(MockBuffer, cl)
	out := []byte(cl)
	return out, nil
}

// ExecWithOpts returns command + args and put these to StdOut.
func (m *MockExecutor) ExecWithOpts(opts *ExecOpts, command string, args ...string) ([]byte, error) {
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
