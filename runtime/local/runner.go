package local

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

func Run(filePath string, cliEnv []string) error {

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

		// create directory structure
		err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm)
		if err != nil {
			return err
		}

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
	}

	// Parse manifest AFTER extraction
	manifestPath := filepath.Join(tempDir, "manifest.yaml")

	var m manifest.Manifest

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("manifest not found: %v", err)
	}

	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	fmt.Printf("DEBUG manifest: %+v\n", m)

	// -----------------------------
	// Build volume args
	// -----------------------------
	var volumeArgs []string

	for _, v := range m.Volumes {
		parts := strings.Split(v, ":")
		if len(parts) == 2 {
			hostPath := parts[0]
			containerPath := parts[1]

			volumeArgs = append(volumeArgs, "-v", hostPath+":"+containerPath)
		}
	}

	// -----------------------------
	// Build env args
	// -----------------------------
	var envArgs []string

	// from manifest
	for _, e := range m.Env {
		val := os.Getenv(e)
		if val != "" {
			envArgs = append(envArgs, "-e", e+"="+val)
		}
	}

	// from CLI
	for _, e := range cliEnv {
		envArgs = append(envArgs, "-e", e)
	}

	// -----------------------------
	// Define container name
	// -----------------------------
	hash, err := getFileHash(filePath)
	if err != nil {
		return err
	}

	containerName := "urx-" + hash

	// remove existing container
	exec.Command("docker", "rm", "-f", containerName).Run()

	// save metadata
	meta := storage.RunMeta{
		ID:        containerName,
		Artifact:  filePath,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	storage.SaveMeta(containerName, meta)

	// -----------------------------
	// Build docker command
	// -----------------------------
	args := []string{
		"run",
		"-d",
		"--name", containerName,
	}

	// add env + volumes
	args = append(args, envArgs...)
	args = append(args, volumeArgs...)

	// decide base image
	image := m.BaseImage
        if image == "" {
            image = "python:3.11"
        }
	// mount app code + run	
	args = append(args,
	   "-v", tempDir+":/workspace",
	   image,
	   "python", "-u", "/workspace/"+m.Entrypoint,
	)

	fmt.Println("DEBUG docker args:", args)

	// run container
	runCmd := exec.Command("docker", args...)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	fmt.Println("[URX] Running container:", containerName)

	err = runCmd.Run()
	if err != nil {
		return err
	}

	fmt.Println("[URX] View logs: urx logs", containerName)

	return nil
}
