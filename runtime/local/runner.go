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
	"time"

	"github.com/vijaythakur89/urx/artifacts/manifest"
	"github.com/vijaythakur89/urx/pkg/events"
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
	// 2. PARSE APPLICATION CONTRACT
	// -----------------------------
	// The canonical contract is always stored at the archive root as urx.yaml.
	application, err := loadApplicationContract(tempDir)
	if err != nil {
		return err
	}

	// -----------------------------
	// 3. RUNTIME CONFIG RESOLUTION
	// -----------------------------
	// Contract v1 supports Python. The image remains a local-executor detail,
	// rather than application-owned contract configuration.
	image := "python:3.11"

	// -----------------------------
	// 5. BUILD ENV VARIABLES
	// -----------------------------
	// Inject environment variables into container.
	// Supports:
	//   - values from system env
	//   - values passed via CLI
	var envArgs []string

	// 1. load .env from project
	envFile := loadEnvFile(filepath.Join(tempDir, "app", ".env"))

	// 2. contract environment requirements
	if application.Spec.Environment != nil {
		for _, e := range application.Spec.Environment.Required {

			// priority: .env → system env
			val := envFile[e]

			if val == "" {
				val = os.Getenv(e)
			}

			if val != "" {
				envArgs = append(envArgs, "-e", e+"="+val)
			}
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
	// Initialize lifecycle event emitter.
	eventEmitter, err := events.NewEmitter(storage.EventFilePath(containerName))
	if err != nil {
		return err
	}

	if err := eventEmitter.Emit(events.Event{
		Type:      events.RunStarted,
		Component: "runtime",
		RunID:     containerName,
		Message:   "URX run started",
	}); err != nil {
		return err
	}

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
	containerPort := 8080
	if application.Spec.Service != nil && application.Spec.Service.ListenPort != 0 {
		containerPort = application.Spec.Service.ListenPort
		exposedPort = containerPort
	} else {
		p, err := getFreePort()
		if err != nil {
			return err
		}
		exposedPort = p
	}

	// build mapping
	portMapping := fmt.Sprintf("%d:%d", exposedPort, containerPort)
	args = append(args, "-p", portMapping)

	// Attach environment configuration.
	args = append(args, envArgs...)
	// -----------------------------
	// 10. MOUNT APPLICATION CODE
	// -----------------------------
	// Mount extracted application payload into the local executor workspace.
	args = append(args,
		"-v", tempDir+":/workspace",
		image,
		"python", "-u", "/workspace/app/"+application.Spec.Entrypoint,
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

		// Container failed to start/run.
		_ = eventEmitter.Emit(events.Event{
			Type:      events.ContainerFailed,
			Component: "docker",
			RunID:     containerName,
			Message:   err.Error(),
		})

		// URX operation failed.
		_ = eventEmitter.Emit(events.Event{
			Type:      events.RunFailed,
			Component: "runtime",
			RunID:     containerName,
			Message:   err.Error(),
		})

		return err
	}

	if err := eventEmitter.Emit(events.Event{
		Type:      events.ContainerStarted,
		Component: "docker",
		RunID:     containerName,
		Message:   "Container started successfully",
	}); err != nil {
		return err
	}
	// Wait for the container to finish and determine the actual result.
	waitCmd := exec.Command("docker", "wait", containerName)

	output, err := waitCmd.Output()
	if err != nil {
		_ = eventEmitter.Emit(events.Event{
			Type:      events.RunFailed,
			Component: "runtime",
			RunID:     containerName,
			Message:   err.Error(),
		})

		return err
	}

	exitCode := strings.TrimSpace(string(output))

	if exitCode != "0" {
		_ = eventEmitter.Emit(events.Event{
			Type:      events.RunFailed,
			Component: "runtime",
			RunID:     containerName,
			Message:   "Container exited with code " + exitCode,
		})

		return fmt.Errorf("container exited with code %s", exitCode)
	}
	// -----------------------------
	// SAVE METADATA (IMPORTANT)
	// -----------------------------
	meta := storage.RunMeta{
		ID:        containerName,
		Artifact:  filePath,
		Timestamp: time.Now().Format(time.RFC3339),
		Port:      exposedPort,
	}

	storage.SaveMeta(containerName, meta)

	// -----------------------------
	// 12. USER GUIDANCE
	// -----------------------------
	fmt.Println("[URX] View logs: urx logs", containerName)

	// -----------------------------
	// 13. URL OUTPUT (deploy only)
	// -----------------------------
	if mode == "deploy" {
		fmt.Println("Service deployed")
		fmt.Printf("URL: http://localhost:%d\n", exposedPort)
	}
	if err := eventEmitter.Emit(events.Event{
		Type:      events.RunCompleted,
		Component: "runtime",
		RunID:     containerName,
		Message:   "URX run completed",
	}); err != nil {
		return err
	}
	return nil
}

func loadApplicationContract(workspace string) (manifest.Application, error) {
	data, err := os.ReadFile(filepath.Join(workspace, "urx.yaml"))
	if err != nil {
		return manifest.Application{}, err
	}

	application, err := manifest.ParseAndValidate(data)
	if err != nil {
		return manifest.Application{}, fmt.Errorf("invalid urx.yaml: %w", err)
	}

	return application, nil
}
