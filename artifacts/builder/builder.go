package builder

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/vijaythakur89/urx/artifacts/manifest"
	"gopkg.in/yaml.v3"
)

const contractFileName = "urx.yaml"

// Build creates a TAR-backed .urx artifact containing one canonical contract
// at urx.yaml and the source application under app/.
func Build(sourceDir string, outputFile string) error {
	sourceAbs, err := filepath.Abs(sourceDir)
	if err != nil {
		return err
	}

	sourceInfo, err := os.Stat(sourceAbs)
	if err != nil {
		return err
	}
	if !sourceInfo.IsDir() {
		return fmt.Errorf("source path must be a directory: %s", sourceDir)
	}

	contractPath := filepath.Join(sourceAbs, contractFileName)
	contractData, err := os.ReadFile(contractPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s is required", contractFileName)
		}
		return err
	}

	application, err := manifest.ParseAndValidate(contractData)
	if err != nil {
		return fmt.Errorf("invalid %s: %w", contractFileName, err)
	}

	if err := validateEntrypoint(sourceAbs, application.Spec.Entrypoint); err != nil {
		return err
	}

	outputAbs, err := filepath.Abs(outputFile)
	if err != nil {
		return err
	}

	canonicalContract, err := yaml.Marshal(application)
	if err != nil {
		return err
	}

	file, err := os.Create(outputAbs)
	if err != nil {
		return err
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	if err := writeContract(tw, canonicalContract); err != nil {
		return err
	}

	err = filepath.Walk(sourceAbs, func(filePath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}

		fileAbs, err := filepath.Abs(filePath)
		if err != nil {
			return err
		}
		if fileAbs == contractPath || fileAbs == outputAbs {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported application file: %s", filePath)
		}

		relativePath, err := filepath.Rel(sourceAbs, fileAbs)
		if err != nil {
			return err
		}
		if !safeRelativePath(relativePath) {
			return fmt.Errorf("unsafe application path: %s", filePath)
		}

		return writeApplicationFile(tw, fileAbs, info, relativePath)
	})
	if err != nil {
		return err
	}

	fmt.Println("URX package created:", outputFile)
	return nil
}

func validateEntrypoint(sourceDir, entrypoint string) error {
	sourceResolved, err := filepath.EvalSymlinks(sourceDir)
	if err != nil {
		return err
	}

	entrypointPath := filepath.Join(sourceDir, filepath.FromSlash(entrypoint))
	entrypointResolved, err := filepath.EvalSymlinks(entrypointPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("spec.entrypoint does not exist: %s", entrypoint)
		}
		return err
	}
	if !isWithin(sourceResolved, entrypointResolved) {
		return fmt.Errorf("spec.entrypoint must remain inside the source directory")
	}

	entrypointInfo, err := os.Stat(entrypointResolved)
	if err != nil {
		return err
	}
	if !entrypointInfo.Mode().IsRegular() {
		return fmt.Errorf("spec.entrypoint must be a regular file: %s", entrypoint)
	}

	return nil
}

func writeContract(tw *tar.Writer, data []byte) error {
	header := &tar.Header{
		Name: contractFileName,
		Mode: 0644,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

func writeApplicationFile(tw *tar.Writer, filePath string, info os.FileInfo, relativePath string) error {
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = path.Join("app", filepath.ToSlash(relativePath))

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(tw, file)
	return err
}

func safeRelativePath(value string) bool {
	if filepath.IsAbs(value) {
		return false
	}

	for _, part := range strings.Split(filepath.ToSlash(value), "/") {
		if part == ".." {
			return false
		}
	}

	cleaned := path.Clean(filepath.ToSlash(value))
	return cleaned != "." && cleaned != ".." && !strings.HasPrefix(cleaned, "../")
}

func isWithin(directory, filePath string) bool {
	relativePath, err := filepath.Rel(directory, filePath)
	if err != nil {
		return false
	}
	return safeRelativePath(relativePath)
}
