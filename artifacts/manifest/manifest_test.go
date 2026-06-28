package manifest

import (
	"testing"

	yamlv3 "gopkg.in/yaml.v3"
)

func TestValidManifest(t *testing.T) {
	data := `
name: demo
runtime: python
entrypoint: app.py
`

	var m Manifest

	err := yamlv3.Unmarshal([]byte(data), &m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Name != "demo" {
		t.Fatalf("expected demo name")
	}

	if m.Runtime != "python" {
		t.Fatalf("expected python runtime")
	}
}

func TestMissingFields(t *testing.T) {
	data := `
runtime: python
`

	var m Manifest

	err := yamlv3.Unmarshal([]byte(data), &m)
	if err != nil {
		t.Fatalf("unexpected yaml error")
	}

	if m.Name == "" {
		// expected behavior
		return
	}

	t.Fatalf("expected missing name")
}

func TestInvalidYAML(t *testing.T) {
	data := `
name: demo
runtime: [python
`

	var m Manifest

	err := yamlv3.Unmarshal([]byte(data), &m)

	if err == nil {
		t.Fatalf("expected yaml parse error")
	}
}
