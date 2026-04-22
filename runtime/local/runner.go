// TODO:
// Add production container runtime:
// - build Docker image
// - push to registry
// - run via image (no volume mount)
package local

import (
	"archive/tar"
	"fmt"
	"io"
	"time"
	"os"
	"os/exec"
	"path/filepath"
        "crypto/sha256"
	"encoding/hex"

	"gopkg.in/yaml.v3"
	"github.com/vijaythakur89/urx/artifacts/manifest"
        "github.com/vijaythakur89/urx/pkg/storage"
)

func getFileHash(filePath string) (string, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(file)
	return hex.EncodeToString(hash[:8]), nil
}

func Run(filePath string) error {

	// create temp dir
	tempDir, err := os.MkdirTemp("", "urx-*")
	if err != nil {
		return err
	}
	fmt.Println("[URX] Extracting to:", tempDir)

	// open tar file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	tr := tar.NewReader(file)

	var m manifest.Manifest

	// extract files
	for {
		header, err := tr.Next()

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(tempDir, header.Name)

		// create file
		f, err := os.Create(targetPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, tr)
		if err != nil {
			return err
		}
		f.Close()

		// parse manifest
		if header.Name == "manifest.yaml" {
			data, err := os.ReadFile(targetPath)
			if err != nil {
				return err
			}

			err = yaml.Unmarshal(data, &m)
			if err != nil {
				return err
			}
		}
	}

	fmt.Println("[URX] Running:", m.Entrypoint)

//Define hash
hash, err := getFileHash(filePath)
if err != nil {
    return err
}

//Running Container
containerName := "urx-" + hash

// remove existing container if exists
exec.Command("docker", "rm", "-f", containerName).Run()

// ADD metadata saving
meta := storage.RunMeta{
	ID:        containerName,
	Artifact:  filePath,
	Timestamp: time.Now().Format(time.RFC3339),
}
storage.SaveMeta(containerName, meta)

// run container
runCmd := exec.Command(
	"docker", "run",
	"-d",
	"--name", containerName,
	"-v", tempDir+":/app",
	"python:3.11",
	"python", "-u", "/app/"+m.Entrypoint,
)

runCmd.Stdout = nil
runCmd.Stderr = os.Stderr

fmt.Println("[URX] Running container:", containerName)

err = runCmd.Run()
if err != nil {
	return err
}

fmt.Println("[URX] View logs: urx logs", containerName)

return nil
}
