package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"alex-hendel.de/gitea-update/internal/config"
	"alex-hendel.de/gitea-update/internal/install"
	"alex-hendel.de/gitea-update/internal/logger"
	"alex-hendel.de/gitea-update/internal/version"
	"github.com/sirupsen/logrus"
)

const Version = "1.0.0"

func main() {
	logger.InitLogger()

	infoFlag := flag.Bool("info", false, "Display version information for gitea and act_runner")
	installFlag := flag.Bool("install", false, "Install the latest version")
	devFlag := flag.Bool("dev", false, "Install the development version if specified")
	pathFlag := flag.String("path", "/opt/gitea", "Path to the gitea and act_runner binaries")
	userFlag := flag.String("user", "app", "User name for file ownership")
	groupFlag := flag.String("group", "app", "Group name for file ownership")
	versionFlag := flag.Bool("version", false, "Prints the version of the program")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	if *verboseFlag {
		logger.Logger.SetLevel(logrus.DebugLevel)
	}

	if *versionFlag {
		fmt.Println("gitea-update version", Version)
		os.Exit(0)
	}

	userConfigDir, _ := os.UserHomeDir()
	configPaths := []string{
		"gitea-update.yml", // current directory
		filepath.Join(userConfigDir, "gitea-update.yml"), // home directory
		"/etc/gitea-update/gitea-update.yml",             // etc directory
	}

	// Load or get default config
	err := config.LoadConfig(configPaths)
	if err != nil {
		logger.LogWarn(fmt.Sprintln("Failed to load config, using default settings:", err))
		config.AppConfig = config.GetDefaultConfig()
	}

	logger.LogInfo("Configuration loaded successfully or defaults in use.")
	logger.LogInfo(fmt.Sprintf("User: %s, Group: %s\n", config.AppConfig.Settings.User, config.AppConfig.Settings.Group))

	if *infoFlag {
		serviceNames := config.ListServices() // Get list of all service names
		for _, serviceName := range serviceNames {
			service, exists := config.GetService(serviceName) // Use a getter to access service details
			if !exists {
				logger.LogWarn(fmt.Sprintf("Service %s not found\n", serviceName))
				continue
			}

			installedVersion := version.CheckInstalledVersion(*pathFlag, service.BinName)
			if installedVersion != "" && installedVersion != service.Version.Current {
				config.UpdateServiceVersion(serviceName, installedVersion) // Update version through a config package method
			}
		}

		install.RetrieveRemoteVersion()
		install.PrintInfo()
		return
	}

	if *installFlag {
		install.RetrieveRemoteVersion()
		install.PerformInstallation(*pathFlag, *devFlag, *userFlag, *groupFlag)
	}
}
