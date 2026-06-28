package storage

import (
	"os"
	"testing"
)

func TestSaveAndLoadMeta(t *testing.T) {

	meta := RunMeta{
		ID:        "test-run",
		Artifact:  "app.urx",
		Timestamp: "2026-01-01T00:00:00Z",
		Port:      8080,
	}

	err := SaveMeta("test-run", meta)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	metas, err := LoadAllMeta()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	found := false

	for _, m := range metas {
		if m.ID == "test-run" {
			found = true

			if m.Port != 8080 {
				t.Fatalf("expected port 8080")
			}
		}
	}

	if !found {
		t.Fatalf("metadata not found")
	}
}

func TestGetRunDir(t *testing.T) {
	dir := GetRunDir("abc")

	if dir == "" {
		t.Fatalf("expected valid directory")
	}
}

func TestLoadAllMetaNoDir(t *testing.T) {

	home, _ := os.UserHomeDir()

	os.RemoveAll(home + "/.urx")

	_, err := LoadAllMeta()

	if err != nil {
		t.Fatalf("expected graceful handling")
	}
}
