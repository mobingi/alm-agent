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

type mockExecuter struct{}

func (m *mockExecuter) Exec(command string, args ...string) ([]byte, error) {
	out := []byte("Mocked")
	return out, nil
}

func TestMockExec(t *testing.T) {
	Executer = &mockExecuter{}
	expected := "Mocked"
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
