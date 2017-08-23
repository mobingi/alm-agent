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
	MockBuffer = nil
}

func TestMockExec(t *testing.T) {
	ClearMockBuffer()
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
	ClearMockBuffer()
}

func TestMockedExecuter(t *testing.T) {
	ClearMockBuffer()
	Executer = &MockExecuter{}
	Executer.Exec("echo", "-n", "Mocked")
	Executer.Exec("/bin/true", "but", "Mocked")

	t.Log(MockBuffer)

	expected := "echo -n Mocked"
	actual := MockBuffer[0]
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}

	expected = "/bin/true but Mocked"
	actual = MockBuffer[1]
	if actual != expected {
		t.Fatalf("Expected: %s\n But: %s", expected, actual)
	}
	ClearMockBuffer()
}
