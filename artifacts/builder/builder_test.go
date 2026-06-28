package builder

import (
	"os"
	"testing"
)

func TestBuildInvalidPath(t *testing.T) {

	err := Build("/invalid/path", "test.urx")

	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestBuildTempProject(t *testing.T) {

	dir := t.TempDir()

	err := os.WriteFile(
		dir+"/app.py",
		[]byte("print('hi')"),
		0644,
	)

	if err != nil {
		t.Fatalf("failed to create app.py")
	}

	manifest := `
name: demo
runtime: python
entrypoint: app.py
`

	err = os.WriteFile(
		dir+"/manifest.yaml",
		[]byte(manifest),
		0644,
	)

	if err != nil {
		t.Fatalf("failed to create manifest")
	}

	output := t.TempDir() + "/app.urx"

	err = Build(dir, output)

	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// verify artifact created
	_, err = os.Stat(output)

	if err != nil {
		t.Fatalf("artifact not created")
	}
}
