package util

import "testing"

func TestRealExec(t *testing.T) {
	expected := "Hello Agent!"
	out, err := Executer.Exec("echo", "-n", expected)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(out))
	actual := string(out)

	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}
}

func TestMockExec(t *testing.T) {
	Executer = &MockExecuter{}
	expected := "echo -n Mocked"
	out, err := Executer.Exec("echo", "-n", "Mocked")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(out))
	actual := string(out)

	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}
}

func TestMockedExecuter(t *testing.T) {
	Executer = &MockExecuter{}
	Executer.Exec("echo", "-n", "Mocked")
	Executer.Exec("/bin/true", "but", "Mocked")

	buf := GetMockBufferr()
	t.Log(buf)

	expected := "echo -n Mocked"
	actual := buf[0]
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}

	expected = "/bin/true but Mocked"
	actual = buf[1]
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}
}
