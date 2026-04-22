package builder

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
)
import (
    
    "gopkg.in/yaml.v3"
    "github.com/vijaythakur89/urx/artifacts/manifest"
)

// Build creates a .urx file from a directory
func Build(sourceDir string, outputFile string) error {

	// create output file
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

        // create default manifest
	m := manifest.Manifest{
	    Name:       filepath.Base(sourceDir),
	    Runtime:    "python",
	    Entrypoint: "app.py",
	    Isolation:  "low",
 }

	// convert to YAML
	data, err := yaml.Marshal(m)
	if err != nil {
	    return err
 }

	// write manifest into tar
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

	// walk through source directory
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip directories
		if info.IsDir() {
			return nil
		}

		// open file
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		// create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// fix name (relative path)
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// write file content
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
