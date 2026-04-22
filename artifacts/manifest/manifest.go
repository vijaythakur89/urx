package manifest

type Manifest struct {
	Name       string `yaml:"name"`
	Runtime    string `yaml:"runtime"`
	BaseImage  string   `yaml:"base_image"`
	Entrypoint string `yaml:"entrypoint"`
	Isolation  string `yaml:"isolation"`
	Port	      int `yaml:"port"`
	Volumes  []string `yaml:"volumes"`
	Env      []string `yaml:"env"`
}
