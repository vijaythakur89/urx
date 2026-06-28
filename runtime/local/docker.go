package local

type DockerClient interface {
	Run(args ...string) error
	Logs(id string) ([]byte, error)
	Inspect(id string) ([]byte, error)
}
