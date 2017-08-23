package util

import (
	"os/exec"
	"strings"
)

// ExecuterInterface has Exec to wrap os commands.
type ExecuterInterface interface {
	Exec(string, ...string) ([]byte, error)
}

type osExecuter struct{}

// Executer is real Executer
var Executer ExecuterInterface

func init() {
	Executer = &osExecuter{}
}

func (o *osExecuter) Exec(command string, args ...string) ([]byte, error) {
	out, err := exec.Command(command, args...).CombinedOutput()
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
	// fmt.Println(cl)
	out := []byte(cl)
	return out, nil
}

// GetMockBufferr returns buffered commands
func GetMockBufferr() []string {
	return MockBuffer
}
