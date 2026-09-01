package manifest

import (
	"strings"
	"testing"
)

const completeContract = `
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

func TestParseAndValidateCompleteContract(t *testing.T) {
	application, err := ParseAndValidate([]byte(completeContract))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if application.APIVersion != "urx/v1" {
		t.Fatalf("expected apiVersion urx/v1, got %q", application.APIVersion)
	}

	if application.Kind != "Application" {
		t.Fatalf("expected kind Application, got %q", application.Kind)
	}

	if application.Metadata.Name != "demo" {
		t.Fatalf("expected metadata name demo, got %q", application.Metadata.Name)
	}

	if application.Spec.Runtime.Name != "python" {
		t.Fatalf("expected runtime name python, got %q", application.Spec.Runtime.Name)
	}

	if application.Spec.Runtime.Version != "3.11" {
		t.Fatalf("expected runtime version 3.11, got %q", application.Spec.Runtime.Version)
	}

	if application.Spec.Entrypoint != "app.py" {
		t.Fatalf("expected entrypoint app.py, got %q", application.Spec.Entrypoint)
	}

	if application.Spec.Service == nil {
		t.Fatal("expected service to be populated")
	}

	if application.Spec.Service.ListenPort != 8080 {
		t.Fatalf("expected listen port 8080, got %d", application.Spec.Service.ListenPort)
	}

	if application.Spec.Environment == nil {
		t.Fatal("expected environment to be populated")
	}

	required := application.Spec.Environment.Required
	if len(required) != 1 || required[0] != "TEST" {
		t.Fatalf("expected required environment [TEST], got %q", required)
	}
}

func TestParseAndValidateMinimalContract(t *testing.T) {
	data := `
apiVersion: urx/v1
kind: Application
metadata:
  name: demo
spec:
  runtime:
    name: python
  entrypoint: app.py
`

	application, err := ParseAndValidate([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if application.Spec.Runtime.Version != "" {
		t.Fatalf("expected omitted runtime version, got %q", application.Spec.Runtime.Version)
	}
	if application.Spec.Service != nil {
		t.Fatal("expected omitted service")
	}
	if application.Spec.Environment != nil {
		t.Fatal("expected omitted environment")
	}
}

func TestParseAndValidateAllowsOptionalNestedFieldsWithoutValues(t *testing.T) {
	data := `
apiVersion: urx/v1
kind: Application
metadata:
  name: demo
spec:
  runtime:
    name: python
  entrypoint: app.py
  service: {}
  environment: {}
`

	application, err := ParseAndValidate([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if application.Spec.Service == nil {
		t.Fatal("expected service to be populated")
	}
	if application.Spec.Service.ListenPort != 0 {
		t.Fatalf("expected omitted listen port, got %d", application.Spec.Service.ListenPort)
	}
	if application.Spec.Environment == nil {
		t.Fatal("expected environment to be populated")
	}
}

func TestParseAndValidateRejectsInvalidContracts(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr string
	}{
		{
			name:    "malformed YAML",
			data:    "apiVersion: urx/v1\nspec: [python\n",
			wantErr: "invalid URX application contract YAML",
		},
		{
			name:    "unknown field",
			data:    strings.Replace(completeContract, "kind: Application", "kind: Application\nbase_image: python:3.11", 1),
			wantErr: "field base_image not found",
		},
		{
			name:    "Docker volume field",
			data:    strings.Replace(completeContract, "  entrypoint: app.py", "  volumes:\n    - /host:/container\n  entrypoint: app.py", 1),
			wantErr: "field volumes not found",
		},
		{
			name:    "Docker isolation field",
			data:    strings.Replace(completeContract, "  entrypoint: app.py", "  isolation: low\n  entrypoint: app.py", 1),
			wantErr: "field isolation not found",
		},
		{
			name:    "misspelled contract field",
			data:    strings.Replace(completeContract, "  entrypoint: app.py", "  entrypiont: app.py", 1),
			wantErr: "field entrypiont not found",
		},
		{
			name:    "wrong apiVersion",
			data:    strings.Replace(completeContract, "apiVersion: urx/v1", "apiVersion: urx/v2", 1),
			wantErr: "invalid apiVersion: expected urx/v1",
		},
		{
			name:    "wrong kind",
			data:    strings.Replace(completeContract, "kind: Application", "kind: Service", 1),
			wantErr: "invalid kind: expected Application",
		},
		{
			name:    "missing name",
			data:    strings.Replace(completeContract, "  name: demo\n", "", 1),
			wantErr: "metadata.name is required",
		},
		{
			name:    "empty name",
			data:    strings.Replace(completeContract, "name: demo", "name: '   '", 1),
			wantErr: "metadata.name is required",
		},
		{
			name:    "missing runtime",
			data:    strings.Replace(completeContract, "  runtime:\n    name: python\n    version: \"3.11\"\n", "", 1),
			wantErr: "spec.runtime.name is required",
		},
		{
			name:    "unsupported runtime",
			data:    strings.Replace(completeContract, "name: python", "name: node", 1),
			wantErr: "unsupported runtime: node",
		},
		{
			name:    "missing entrypoint",
			data:    strings.Replace(completeContract, "  entrypoint: app.py\n", "", 1),
			wantErr: "spec.entrypoint is required",
		},
		{
			name:    "absolute entrypoint",
			data:    strings.Replace(completeContract, "entrypoint: app.py", "entrypoint: /app.py", 1),
			wantErr: "spec.entrypoint must be a relative path",
		},
		{
			name:    "traversal entrypoint",
			data:    strings.Replace(completeContract, "entrypoint: app.py", "entrypoint: ../app.py", 1),
			wantErr: "spec.entrypoint must be a relative path",
		},
		{
			name:    "port zero when supplied",
			data:    strings.Replace(completeContract, "listen_port: 8080", "listen_port: 0", 1),
			wantErr: "spec.service.listen_port must be between 1 and 65535",
		},
		{
			name:    "port above range",
			data:    strings.Replace(completeContract, "listen_port: 8080", "listen_port: 65536", 1),
			wantErr: "spec.service.listen_port must be between 1 and 65535",
		},
		{
			name:    "invalid environment variable",
			data:    strings.Replace(completeContract, "- TEST", "- INVALID-NAME", 1),
			wantErr: "invalid required environment variable: INVALID-NAME",
		},
		{
			name:    "duplicate environment variable",
			data:    strings.Replace(completeContract, "      - TEST", "      - TEST\n      - TEST", 1),
			wantErr: "duplicate required environment variable: TEST",
		},
		{
			name:    "explicit empty runtime version",
			data:    strings.Replace(completeContract, "version: \"3.11\"", "version: \"\"", 1),
			wantErr: "spec.runtime.version must not be empty when supplied",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseAndValidate([]byte(test.data))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("expected error containing %q, got %q", test.wantErr, err)
			}
		})
	}
}
