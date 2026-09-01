package inspector

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vijaythakur89/urx/artifacts/manifest"
)

const contractFileName = "urx.yaml"

// Inspect reads and displays the canonical URX Application Contract stored in
// an artifact.
func Inspect(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	contractData, err := readContract(tar.NewReader(file))
	if err != nil {
		return err
	}

	application, err := manifest.ParseAndValidate(contractData)
	if err != nil {
		return fmt.Errorf("invalid %s: %w", contractFileName, err)
	}

	printApplication(application)
	return nil
}

func readContract(reader *tar.Reader) ([]byte, error) {
	var contractData []byte
	contractCount := 0

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Name != contractFileName {
			continue
		}

		contractCount++
		if contractCount > 1 {
			return nil, fmt.Errorf("artifact contains multiple %s entries", contractFileName)
		}

		contractData, err = io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
	}

	if contractCount == 0 {
		return nil, fmt.Errorf("%s not found", contractFileName)
	}

	return contractData, nil
}

func printApplication(application manifest.Application) {
	fmt.Println("URX Application Contract:")
	fmt.Println("-------------------------")
	fmt.Println("apiVersion:", application.APIVersion)
	fmt.Println("kind:", application.Kind)
	fmt.Println("name:", application.Metadata.Name)
	fmt.Println("runtime:", application.Spec.Runtime.Name)
	if application.Spec.Runtime.Version != "" {
		fmt.Println("runtime version:", application.Spec.Runtime.Version)
	}
	fmt.Println("entrypoint:", application.Spec.Entrypoint)

	if application.Spec.Service != nil && application.Spec.Service.ListenPort != 0 {
		fmt.Println("listen port:", application.Spec.Service.ListenPort)
	}
	if application.Spec.Environment != nil && len(application.Spec.Environment.Required) > 0 {
		fmt.Println("required environment:", strings.Join(application.Spec.Environment.Required, ", "))
	}
}
