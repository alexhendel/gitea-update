package config

type Config struct {
	Settings struct {
		User     string             `yaml:"user"`
		Group    string             `yaml:"group"`
		Services map[string]Service `yaml:"services"`
	} `yaml:"settings"`
}

type Service struct {
	BinName string         `yaml:"bin"`
	Path    string         `yaml:"path"`
	URLs    ServiceURLs    `yaml:"urls"`
	Version ServiceVersion `yaml:"version"`
}

type ServiceURLs struct {
	Download string `yaml:"download"`
	API      string `yaml:"api"`
}

type ServiceVersion struct {
	Current string `yaml:"current"`
	Latest  string `yaml:"latest"`
	Dev     string `yaml:"dev"`
}
