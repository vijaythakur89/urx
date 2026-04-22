package manifest

type Manifest struct {
	Name       string `yaml:"name"`
	Runtime    string `yaml:"runtime"`
	Entrypoint string `yaml:"entrypoint"`
	Isolation  string `yaml:"isolation"`
}
