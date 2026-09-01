package local

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestLoadApplicationContractReadsCanonicalRootContract(t *testing.T) {
	workspace := t.TempDir()
	contract := `
apiVersion: urx/v1
kind: Application
metadata:
  name: demo
spec:
  runtime:
    name: python
  entrypoint: app.py
`

	if err := os.WriteFile(filepath.Join(workspace, "urx.yaml"), []byte(contract), 0644); err != nil {
		t.Fatalf("failed to write contract: %v", err)
	}
	if err := os.Mkdir(filepath.Join(workspace, "app"), 0755); err != nil {
		t.Fatalf("failed to create payload directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "app", "urx.yaml"), []byte("base_image: ignored"), 0644); err != nil {
		t.Fatalf("failed to write nested file: %v", err)
	}

	application, err := loadApplicationContract(workspace)
	if err != nil {
		t.Fatalf("failed to load contract: %v", err)
	}
	if application.Metadata.Name != "demo" {
		t.Fatalf("expected root contract name demo, got %q", application.Metadata.Name)
	}
}
