package config

import (
	"fmt"
	"os"
	"sync"

	"alex-hendel.de/gitea-update/internal/logger"
	"gopkg.in/yaml.v2"
)

var AppConfig *Config
var mutex sync.Mutex

func ListServices() []string {
	var serviceNames []string
	if AppConfig == nil || AppConfig.Settings.Services == nil {
		return serviceNames // Return an empty slice if not initialized
	}
	for name := range AppConfig.Settings.Services {
		serviceNames = append(serviceNames, name)
	}
	return serviceNames
}

func GetService(name string) (Service, bool) {
	service, exists := AppConfig.Settings.Services[name]
	return service, exists
}

func UpdateService(name string, service *Service) {
	mutex.Lock()
	defer mutex.Unlock()

	logger.LogDebug(fmt.Sprintf("Updating service %s: %s", name, service))

	if _, exists := AppConfig.Settings.Services[name]; exists {
		AppConfig.Settings.Services[name] = *service // Update the struct in the map
		logger.LogDebug(fmt.Sprintf("Service %s exists, updating %s", name, AppConfig.Settings.Services[name]))
	}
}

func UpdateServiceVersion(name, version string) {
	if service, exists := AppConfig.Settings.Services[name]; exists {
		service.Version.Current = version
		AppConfig.Settings.Services[name] = service
	}
}

func LoadConfig(paths []string) error {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			yamlFile, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read YAML file: %s", err)
			}
			if err = yaml.Unmarshal(yamlFile, &AppConfig); err != nil {
				return fmt.Errorf("failed to unmarshal YAML file: %s", err)
			}
			return nil
		}
	}
	return fmt.Errorf("config file not found in provided paths")
}

func GetDefaultConfig() *Config {
	return &Config{
		Settings: struct {
			User     string             `yaml:"user"`
			Group    string             `yaml:"group"`
			Services map[string]Service `yaml:"services"`
		}{
			User:  "app",
			Group: "app",
			Services: map[string]Service{
				"gitea": {
					BinName: "gitea",
					Path:    "/opt/gitea",
					URLs: ServiceURLs{
						Download: "https://dl.gitea.io/gitea/{version}/gitea-{version}-linux-amd64",
						API:      "https://api.github.com/repos/go-gitea/gitea/tags",
					},
					Version: ServiceVersion{
						Current: "n/a",
						Latest:  "n/a",
						Dev:     "n/a",
					},
				},
				"act_runner": {
					BinName: "act_runner",
					Path:    "/opt/gitea",
					URLs: ServiceURLs{
						Download: "https://dl.gitea.com/act_runner/{version}/act_runner-{version}-linux-amd64",
						API:      "https://gitea.com/api/v1/repos/gitea/act_runner/tags",
					},
					Version: ServiceVersion{
						Current: "n/a",
						Latest:  "n/a",
						Dev:     "n/a",
					},
				},
			},
		},
	}
}
