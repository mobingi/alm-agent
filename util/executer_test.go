package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRealExec(t *testing.T) {
	defer ClearMockBuffer()

	assert := assert.New(t)

	expected := "Hello Agent!"
	out, err := Executer.Exec("echo", "-n", expected)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(string(out), expected)
}

func TestMockExec(t *testing.T) {
	defer ClearMockBuffer()

	assert := assert.New(t)

	Executer = &MockExecuter{}
	expected := "echo -n Mocked"
	out, err := Executer.Exec("echo", "-n", "Mocked")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(string(out), expected)
}

func TestMockedExecuter(t *testing.T) {
	defer ClearMockBuffer()

	assert := assert.New(t)

	Executer = &MockExecuter{}
	Executer.Exec("echo", "-n", "Mocked")
	Executer.Exec("/bin/true", "but", "Mocked")

	assert.Equal(MockBuffer[0], "echo -n Mocked")
	assert.Equal(MockBuffer[1], "/bin/true but Mocked")
}
