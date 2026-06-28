package local

import "testing"

func TestMockDockerLogs(t *testing.T) {

	mock := &MockDockerClient{}

	out, err := mock.Logs("abc")

	if err != nil {
		t.Fatalf("unexpected error")
	}

	if string(out) != "mock logs" {
		t.Fatalf("unexpected logs")
	}
}
