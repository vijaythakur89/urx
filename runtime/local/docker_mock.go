package local

type MockDockerClient struct{}

func (m *MockDockerClient) Run(args ...string) error {
	return nil
}

func (m *MockDockerClient) Logs(id string) ([]byte, error) {
	return []byte("mock logs"), nil
}

func (m *MockDockerClient) Inspect(id string) ([]byte, error) {
	return []byte("running"), nil
}
func (m *MockDockerClient) Wait(id string) (string, error) {
	return "0", nil
}

func (m *MockDockerClient) Remove(id string) error {
	return nil
}
