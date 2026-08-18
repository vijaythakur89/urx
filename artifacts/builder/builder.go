package builder

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/vijaythakur89/urx/artifacts/manifest"
	"gopkg.in/yaml.v3"
)

// Build creates a .urx file from a directory.
func Build(sourceDir string, outputFile string) error {

	// Resolve absolute paths so we can safely identify the output artifact
	// while walking the source directory.
	sourceAbs, err := filepath.Abs(sourceDir)
	if err != nil {
		return err
	}

	outputAbs, err := filepath.Abs(outputFile)
	if err != nil {
		return err
	}

	// Create the output artifact.
	file, err := os.Create(outputAbs)
	if err != nil {
		return err
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	// Create default manifest.
	m := manifest.Manifest{
		Name:       filepath.Base(sourceAbs),
		Runtime:    "python",
		Entrypoint: "app.py",
		Isolation:  "low",
	}

	// Convert manifest to YAML.
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	// Write manifest into tar.
	header := &tar.Header{
		Name: "manifest.yaml",
		Mode: 0600,
		Size: int64(len(data)),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := tw.Write(data); err != nil {
		return err
	}

	// Walk through source directory.
	err = filepath.Walk(sourceAbs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories.
		if info.IsDir() {
			return nil
		}

		// Never package the output artifact itself.
		pathAbs, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		if pathAbs == outputAbs {
			return nil
		}

		// Open file.
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		// Create tar header.
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Use a path relative to the source directory.
		relPath, err := filepath.Rel(sourceAbs, pathAbs)
		if err != nil {
			return err
		}

		header.Name = relPath

		// Write header.
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file content.
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	fmt.Println("URX package created:", outputFile)
	return nil
}
