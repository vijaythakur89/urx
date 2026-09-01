package local

type DockerClient interface {
	Run(args ...string) error
	Wait(id string) (string, error)
	Logs(id string) ([]byte, error)
	Inspect(id string) ([]byte, error)
	Remove(id string) error
}
