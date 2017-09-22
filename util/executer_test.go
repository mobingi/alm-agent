package util

import (
	"io/ioutil"
	"os"
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
func TestRealExecWithOpts(t *testing.T) {
	defer ClearMockBuffer()
	tmpDir, _ := ioutil.TempDir("", "exec")
	defer os.RemoveAll(tmpDir)

	assert := assert.New(t)

	opts := &ExecOpts{
		Dir: tmpDir,
	}
	out, err := Executer.ExecWithOpts(opts, "pwd")
	if err != nil {
		t.Fatal(err)
	}

	// `/private/` hides from shell on macos.
	assert.Contains(string(out), tmpDir)
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

func TestMockedExecuterWithOpts(t *testing.T) {
	defer ClearMockBuffer()

	assert := assert.New(t)

	Executer = &MockExecuter{}

	env := []string{"a=1", "b=2"}
	opts := &ExecOpts{
		Dir: "/tmp",
		Env: env,
	}

	Executer.ExecWithOpts(opts, "echo", "-n", "Mocked")

	assert.Equal(MockBuffer[0], "echo -n Mocked")
	assert.Equal(MockBuffer[1], "/tmp")
	assert.Equal(MockBuffer[2], "a=1,b=2")
}
