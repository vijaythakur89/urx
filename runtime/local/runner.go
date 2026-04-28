package local

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
        "net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"gopkg.in/yaml.v3"

	"github.com/vijaythakur89/urx/artifacts/manifest"
)

func getFileHash(filePath string) (string, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(file)
	return hex.EncodeToString(hash[:8]), nil
}
func getFreePort() (int, error) {
        listener, err := net.Listen("tcp", ":0")
        if err != nil {
                return 0, err
        }
        defer listener.Close()

        addr := listener.Addr().(*net.TCPAddr)
        return addr.Port, nil
}

func loadEnvFile(path string) map[string]string {
	envs := make(map[string]string)

	data, err := os.ReadFile(path)
	if err != nil {
		return envs // ignore if not present
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envs[parts[0]] = parts[1]
		}
	}

	return envs
}

func Run(filePath string, cliEnv []string) error {
    return RunWithMode(filePath, "run", cliEnv)
}


func Deploy(filePath string) error {
	return RunWithMode(filePath, "deploy", nil)
}

func RunWithMode(filePath string, mode string, cliEnv []string) error {

	// -----------------------------
	// 1. EXTRACT ARTIFACT
	// -----------------------------
	// Create a temp workspace where the .urx file will be unpacked.
	// This simulates a filesystem inside the container.
	tempDir, err := os.MkdirTemp("", "urx-*")
	if err != nil {
		return err
	}
	fmt.Println("[URX] Extracting to:", tempDir)

	// Open the .urx file (tar archive)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	tr := tar.NewReader(file)

	var m manifest.Manifest

	// Loop through all files inside the tar archive
	for {
		header, err := tr.Next()

		if err == io.EOF {
			break // no more files
		}
		if err != nil {
			return err
		}

		// Construct full path where file will be written
		targetPath := filepath.Join(tempDir, header.Name)

		// Ensure directory structure exists before writing file
		os.MkdirAll(filepath.Dir(targetPath), os.ModePerm)

		// Create file and copy content from tar
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

	// -----------------------------
	// 2. PARSE MANIFEST
	// -----------------------------
	// Read manifest.yaml AFTER extraction (critical fix)
	// This defines how the app should run.
	data, err := os.ReadFile(filepath.Join(tempDir, "manifest.yaml"))
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	// -----------------------------
	// 3. RUNTIME CONFIG RESOLUTION
	// -----------------------------
	// Decide which container image to use.
	// If user didn't specify, fallback to default runtime.
	image := m.BaseImage
	if image == "" {
		image = "python:3.11"
	}

	// -----------------------------
	// 4. BUILD VOLUME MOUNTS
	// -----------------------------
	// Convert manifest volumes into docker "-v" flags.
	// Example: "/host:/container"
	var volumeArgs []string
	for _, v := range m.Volumes {
		volumeArgs = append(volumeArgs, "-v", v)
	}

	// -----------------------------
	// 5. BUILD ENV VARIABLES
	// -----------------------------
	// Inject environment variables into container.
	// Supports:
	//   - values from system env
	//   - values passed via CLI
	var envArgs []string

	// 1. load .env from project
	envFile := loadEnvFile(filepath.Join(tempDir, ".env"))

	// 2. manifest env
	for _, e := range m.Env {

	// priority: .env → system env
	val := envFile[e]

	if val == "" {
		val = os.Getenv(e)
	}

	if val != "" {
		envArgs = append(envArgs, "-e", e+"="+val)
	}
	}

	// 3. CLI env (highest priority)
	for _, e := range cliEnv {
	envArgs = append(envArgs, "-e", e)
	}

	// -----------------------------
	// 6. GENERATE UNIQUE CONTAINER NAME
	// -----------------------------
	// Hash of artifact ensures deterministic naming.
	hash, err := getFileHash(filePath)
	if err != nil {
		return err
	}
	containerName := "urx-" + hash

	// Remove existing container with same name (idempotent behavior)
	exec.Command("docker", "rm", "-f", containerName).Run()

	// -----------------------------
	// 7. BUILD DOCKER COMMAND
	// -----------------------------
	// Base docker run command
	args := []string{
		"run",
		"-d", // detached mode
		"--name", containerName,
	}

	// -----------------------------
	// 8. DEPLOY MODE BEHAVIOR
	// -----------------------------
	// In deploy mode, container behaves like a service.
	if mode == "deploy" {
		args = append(args, "--restart", "unless-stopped")
	}

	// -----------------------------
	// 9. PORT EXPOSURE
	// -----------------------------
	var exposedPort int

	// decide host port
	if m.Port != 0 {
		exposedPort = m.Port
	} else {
		p, err := getFreePort()
		if err != nil {
			return err
		}
		exposedPort = p
	}

	// container internal port (app listens here)
	containerPort := m.Port
	if containerPort == 0 {
		containerPort = 8080
	}

	// build mapping
	portMapping := fmt.Sprintf("%d:%d", exposedPort, containerPort)
	args = append(args, "-p", portMapping)

	// Attach env + volumes
	args = append(args, envArgs...)
	args = append(args, volumeArgs...)
	// -----------------------------
	// 10. MOUNT APPLICATION CODE
	// -----------------------------
	// Mount extracted workspace into container.
	args = append(args,
		"-v", tempDir+":/workspace",
		image,
		"python", "-u", "/workspace/"+m.Entrypoint,
	)

	//DOcker Debug Mode //fmt.Println("DEBUG docker args:", args)
	// -----------------------------
	// 11. EXECUTE CONTAINER
	// -----------------------------
	runCmd := exec.Command("docker", args...)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	fmt.Println("[URX] Running container:", containerName)

	err = runCmd.Run()
	if err != nil {
    		return err
	}

	// -----------------------------
	// 12. USER GUIDANCE
	// -----------------------------
	fmt.Println("[URX] View logs: urx logs", containerName)

	// -----------------------------
	// 13. URL OUTPUT (deploy only)
	// -----------------------------
	if mode == "deploy" {
	    fmt.Println("🚀 Service deployed")
	    fmt.Printf("URL: http://localhost:%d\n", exposedPort)
	}

	return nil
	}
