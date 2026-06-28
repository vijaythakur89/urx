package local

import "os/exec"

type RealDockerClient struct{}

func (d *RealDockerClient) Run(args ...string) error {

	cmd := exec.Command("docker", args...)

	return cmd.Run()
}

func (d *RealDockerClient) Logs(id string) ([]byte, error) {

	return exec.Command("docker", "logs", id).Output()
}

func (d *RealDockerClient) Inspect(id string) ([]byte, error) {

	return exec.Command("docker", "inspect", id).Output()
}
