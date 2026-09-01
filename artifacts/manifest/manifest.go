package manifest

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Application is the URX Application Contract v1.
// It describes application intent independently of an execution backend.
type Application struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type Metadata struct {
	Name string `yaml:"name"`
}

type Spec struct {
	Runtime     Runtime      `yaml:"runtime"`
	Entrypoint  string       `yaml:"entrypoint"`
	Service     *Service     `yaml:"service,omitempty"`
	Environment *Environment `yaml:"environment,omitempty"`
}

type Runtime struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version,omitempty"`
}

type Service struct {
	ListenPort int `yaml:"listen_port,omitempty"`
}

type Environment struct {
	Required []string `yaml:"required,omitempty"`
}

var environmentVariableName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ParseAndValidate strictly parses a URX Application Contract and validates
// its Contract v1 structure and semantics.
func ParseAndValidate(data []byte) (Application, error) {
	application, err := parse(data)
	if err != nil {
		return Application{}, err
	}

	if err := validateStructure(application, listenPortWasProvided(data)); err != nil {
		return Application{}, err
	}

	if err := validateSemantics(application); err != nil {
		return Application{}, err
	}

	if runtimeVersionWasProvided(data) && strings.TrimSpace(application.Spec.Runtime.Version) == "" {
		return Application{}, fmt.Errorf("spec.runtime.version must not be empty when supplied")
	}

	return application, nil
}

// parse is intentionally separate from validation so malformed YAML and an
// invalid contract can be reported as distinct user errors.
func parse(data []byte) (Application, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var application Application
	if err := decoder.Decode(&application); err != nil {
		return Application{}, fmt.Errorf("invalid URX application contract YAML: %w", err)
	}

	var additionalDocument yaml.Node
	if err := decoder.Decode(&additionalDocument); err != io.EOF {
		if err == nil {
			return Application{}, fmt.Errorf("invalid URX application contract YAML: multiple documents are not supported")
		}
		return Application{}, fmt.Errorf("invalid URX application contract YAML: %w", err)
	}

	return application, nil
}

func validateStructure(application Application, listenPortProvided bool) error {
	if application.APIVersion != "urx/v1" {
		return fmt.Errorf("invalid apiVersion: expected urx/v1")
	}

	if application.Kind != "Application" {
		return fmt.Errorf("invalid kind: expected Application")
	}

	if strings.TrimSpace(application.Metadata.Name) == "" {
		return fmt.Errorf("metadata.name is required")
	}

	if strings.TrimSpace(application.Spec.Runtime.Name) == "" {
		return fmt.Errorf("spec.runtime.name is required")
	}

	if strings.TrimSpace(application.Spec.Entrypoint) == "" {
		return fmt.Errorf("spec.entrypoint is required")
	}

	if !isRelativeApplicationPath(application.Spec.Entrypoint) {
		return fmt.Errorf("spec.entrypoint must be a relative path")
	}

	if listenPortProvided {
		port := application.Spec.Service.ListenPort
		if port < 1 || port > 65535 {
			return fmt.Errorf("spec.service.listen_port must be between 1 and 65535")
		}
	}

	if application.Spec.Environment != nil {
		seen := make(map[string]struct{})
		for _, name := range application.Spec.Environment.Required {
			trimmedName := strings.TrimSpace(name)
			if trimmedName == "" {
				return fmt.Errorf("required environment variable name must not be empty")
			}
			if !environmentVariableName.MatchString(name) {
				return fmt.Errorf("invalid required environment variable: %s", name)
			}
			if _, exists := seen[trimmedName]; exists {
				return fmt.Errorf("duplicate required environment variable: %s", trimmedName)
			}
			seen[trimmedName] = struct{}{}
		}
	}

	return nil
}

func validateSemantics(application Application) error {
	if application.Spec.Runtime.Name != "python" {
		return fmt.Errorf("unsupported runtime: %s", application.Spec.Runtime.Name)
	}

	return nil
}

func isRelativeApplicationPath(value string) bool {
	if strings.HasPrefix(value, "/") || path.IsAbs(value) {
		return false
	}

	for _, part := range strings.Split(value, "/") {
		if part == ".." {
			return false
		}
	}

	cleaned := path.Clean(value)
	return cleaned != "." && cleaned != ".." && !strings.HasPrefix(cleaned, "../")
}

// runtimeVersionWasProvided distinguishes an omitted version from an explicit
// empty YAML value while retaining Version as a simple optional string.
func runtimeVersionWasProvided(data []byte) bool {
	var document yaml.Node
	if yaml.Unmarshal(data, &document) != nil || len(document.Content) == 0 {
		return false
	}

	root := document.Content[0]
	spec := mappingValue(root, "spec")
	runtime := mappingValue(spec, "runtime")
	return mappingValue(runtime, "version") != nil
}

func listenPortWasProvided(data []byte) bool {
	var document yaml.Node
	if yaml.Unmarshal(data, &document) != nil || len(document.Content) == 0 {
		return false
	}

	root := document.Content[0]
	spec := mappingValue(root, "spec")
	service := mappingValue(spec, "service")
	return mappingValue(service, "listen_port") != nil
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == key {
			return node.Content[index+1]
		}
	}

	return nil
}
