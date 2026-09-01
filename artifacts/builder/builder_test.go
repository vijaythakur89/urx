package builder

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vijaythakur89/urx/artifacts/manifest"
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

func TestBuildCreatesCanonicalArtifact(t *testing.T) {
	sourceDir := t.TempDir()
	writeFile(t, filepath.Join(sourceDir, "urx.yaml"), validContract)
	writeFile(t, filepath.Join(sourceDir, "app.py"), "print('hi')")
	writeFile(t, filepath.Join(sourceDir, "data", "input.txt"), "input")

	output := filepath.Join(t.TempDir(), "app.urx")
	if err := Build(sourceDir, output); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	entries := readArtifact(t, output)
	if len(entries[contractFileName]) != 1 {
		t.Fatalf("expected exactly one %s entry, got %d", contractFileName, len(entries[contractFileName]))
	}
	if _, ok := entries["app/app.py"]; !ok {
		t.Fatal("expected app/app.py in artifact")
	}
	if _, ok := entries["app/data/input.txt"]; !ok {
		t.Fatal("expected nested application file in artifact")
	}
	if _, ok := entries["app/urx.yaml"]; ok {
		t.Fatal("source urx.yaml must not be included under app/")
	}

	if _, err := manifest.ParseAndValidate(entries[contractFileName][0]); err != nil {
		t.Fatalf("stored contract is invalid: %v", err)
	}
}

func TestBuildRejectsMissingEntrypoint(t *testing.T) {
	sourceDir := t.TempDir()
	writeFile(t, filepath.Join(sourceDir, "urx.yaml"), validContract)

	err := Build(sourceDir, filepath.Join(t.TempDir(), "app.urx"))
	if err == nil || !strings.Contains(err.Error(), "spec.entrypoint does not exist") {
		t.Fatalf("expected missing entrypoint error, got %v", err)
	}
}

func TestBuildRejectsMissingContract(t *testing.T) {
	sourceDir := t.TempDir()
	writeFile(t, filepath.Join(sourceDir, "app.py"), "print('hi')")

	err := Build(sourceDir, filepath.Join(t.TempDir(), "app.urx"))
	if err == nil || !strings.Contains(err.Error(), "urx.yaml is required") {
		t.Fatalf("expected missing contract error, got %v", err)
	}
}

func TestBuildRejectsInvalidContract(t *testing.T) {
	tests := []struct {
		name     string
		contract string
		wantErr  string
	}{
		{
			name:     "invalid contract",
			contract: strings.Replace(validContract, "kind: Application", "kind: Service", 1),
			wantErr:  "invalid kind: expected Application",
		},
		{
			name:     "Docker base image",
			contract: strings.Replace(validContract, "kind: Application", "kind: Application\nbase_image: python:3.11", 1),
			wantErr:  "field base_image not found",
		},
		{
			name:     "Docker volume",
			contract: strings.Replace(validContract, "  entrypoint: app.py", "  volumes:\n    - /host:/container\n  entrypoint: app.py", 1),
			wantErr:  "field volumes not found",
		},
		{
			name:     "Docker isolation",
			contract: strings.Replace(validContract, "  entrypoint: app.py", "  isolation: low\n  entrypoint: app.py", 1),
			wantErr:  "field isolation not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceDir := t.TempDir()
			writeFile(t, filepath.Join(sourceDir, "urx.yaml"), test.contract)
			writeFile(t, filepath.Join(sourceDir, "app.py"), "print('hi')")

			err := Build(sourceDir, filepath.Join(t.TempDir(), "app.urx"))
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("expected error containing %q, got %v", test.wantErr, err)
			}
		})
	}
}

func TestBuildDoesNotPackageOutputInsideSourceDirectory(t *testing.T) {
	sourceDir := t.TempDir()
	writeFile(t, filepath.Join(sourceDir, "urx.yaml"), validContract)
	writeFile(t, filepath.Join(sourceDir, "app.py"), "print('hi')")

	output := filepath.Join(sourceDir, "app.urx")
	if err := Build(sourceDir, output); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	entries := readArtifact(t, output)
	if _, ok := entries["app/app.urx"]; ok {
		t.Fatal("output artifact must not be packaged recursively")
	}
}

func TestBuildInvalidPath(t *testing.T) {
	if err := Build("/invalid/path", filepath.Join(t.TempDir(), "app.urx")); err == nil {
		t.Fatal("expected error")
	}
}

func writeFile(t *testing.T, filePath, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(contents), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}

func readArtifact(t *testing.T, filePath string) map[string][][]byte {
	t.Helper()

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open artifact: %v", err)
	}
	defer file.Close()

	entries := make(map[string][][]byte)
	reader := tar.NewReader(file)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read artifact: %v", err)
		}

		contents, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to read artifact entry: %v", err)
		}
		entries[header.Name] = append(entries[header.Name], contents)
	}

	return entries
}
