package inspector

import (
	"archive/tar"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
	"github.com/vijaythakur89/urx/artifacts/manifest"
)

func Inspect(filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	tr := tar.NewReader(file)

	for {
		header, err := tr.Next()

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// look for manifest.yaml
		if header.Name == "manifest.yaml" {

			data, err := io.ReadAll(tr)
			if err != nil {
				return err
			}

			var m manifest.Manifest
			err = yaml.Unmarshal(data, &m)
			if err != nil {
				return err
			}

			// pretty print
			fmt.Println("URX Manifest:")
			fmt.Println("-------------")
			fmt.Println("Name       :", m.Name)
			fmt.Println("Runtime    :", m.Runtime)
			fmt.Println("Entrypoint :", m.Entrypoint)
			fmt.Println("Isolation  :", m.Isolation)

			return nil
		}
	}

	return fmt.Errorf("manifest.yaml not found")
}
