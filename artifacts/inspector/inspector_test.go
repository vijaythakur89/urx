package inspector

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validContract = `
apiVersion: urx/v1
kind: Application
metadata:
  name: demo
spec:
  runtime:
    name: python
    version: "3.11"
  entrypoint: app.py
  service:
    listen_port: 8080
  environment:
    required:
      - TEST
`

func TestInspectReadsRootContractAndDisplaysFields(t *testing.T) {
	artifact := writeArtifact(t, map[string]string{
		contractFileName: validContract,
		"app/app.py":     "print('hello')",
		"app/urx.yaml":   "base_image: should-not-be-read",
	})

	output, err := captureOutput(func() error {
		return Inspect(artifact)
	})
	if err != nil {
		t.Fatalf("inspect failed: %v", err)
	}

	for _, field := range []string{
		"apiVersion: urx/v1",
		"kind: Application",
		"name: demo",
		"runtime: python",
		"runtime version: 3.11",
		"entrypoint: app.py",
		"listen port: 8080",
		"required environment: TEST",
	} {
		if !strings.Contains(output, field) {
			t.Fatalf("expected output to contain %q, got %q", field, output)
		}
	}
}

func TestInspectRejectsMissingRootContract(t *testing.T) {
	artifact := writeArtifact(t, map[string]string{
		"app/urx.yaml": validContract,
		"app/app.py":   "print('hello')",
	})

	err := Inspect(artifact)
	if err == nil || !strings.Contains(err.Error(), "urx.yaml not found") {
		t.Fatalf("expected missing root contract error, got %v", err)
	}
}

func TestInspectRejectsDuplicateRootContract(t *testing.T) {
	artifact := writeArtifactEntries(t, []artifactEntry{
		{name: contractFileName, contents: validContract},
		{name: "app/app.py", contents: "print('hello')"},
		{name: contractFileName, contents: validContract},
	})

	err := Inspect(artifact)
	if err == nil || !strings.Contains(err.Error(), "multiple urx.yaml entries") {
		t.Fatalf("expected duplicate contract error, got %v", err)
	}
}

func TestInspectRejectsInvalidContract(t *testing.T) {
	artifact := writeArtifact(t, map[string]string{
		contractFileName: strings.Replace(validContract, "kind: Application", "kind: Service", 1),
	})

	err := Inspect(artifact)
	if err == nil || !strings.Contains(err.Error(), "invalid kind: expected Application") {
		t.Fatalf("expected invalid contract error, got %v", err)
	}
}

func TestInspectRejectsDockerSpecificContractFields(t *testing.T) {
	artifact := writeArtifact(t, map[string]string{
		contractFileName: strings.Replace(validContract, "kind: Application", "kind: Application\nbase_image: python:3.11", 1),
	})

	err := Inspect(artifact)
	if err == nil || !strings.Contains(err.Error(), "field base_image not found") {
		t.Fatalf("expected strict parsing error, got %v", err)
	}
}

type artifactEntry struct {
	name     string
	contents string
}

func writeArtifact(t *testing.T, entries map[string]string) string {
	t.Helper()

	orderedEntries := make([]artifactEntry, 0, len(entries))
	for name, contents := range entries {
		orderedEntries = append(orderedEntries, artifactEntry{name: name, contents: contents})
	}
	return writeArtifactEntries(t, orderedEntries)
}

func writeArtifactEntries(t *testing.T, entries []artifactEntry) string {
	t.Helper()

	artifact := filepath.Join(t.TempDir(), "app.urx")
	file, err := os.Create(artifact)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	writer := tar.NewWriter(file)
	for _, entry := range entries {
		if err := writer.WriteHeader(&tar.Header{Name: entry.name, Mode: 0644, Size: int64(len(entry.contents))}); err != nil {
			t.Fatalf("failed to write archive header: %v", err)
		}
		if _, err := writer.Write([]byte(entry.contents)); err != nil {
			t.Fatalf("failed to write archive content: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close archive writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close artifact: %v", err)
	}

	return artifact
}

func captureOutput(run func() error) (string, error) {
	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		return "", err
	}

	os.Stdout = writer
	runErr := run()
	writer.Close()
	os.Stdout = originalStdout

	output, readErr := io.ReadAll(reader)
	reader.Close()
	if runErr != nil {
		return "", runErr
	}
	if readErr != nil {
		return "", readErr
	}

	return string(bytes.TrimSpace(output)), nil
}
