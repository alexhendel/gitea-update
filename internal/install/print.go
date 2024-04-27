package install

import (
	"fmt"

	"alex-hendel.de/gitea-update/internal/config"
	"alex-hendel.de/gitea-update/internal/logger"
	"alex-hendel.de/gitea-update/internal/version"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

func PrintInfo() {
	for _, service := range config.AppConfig.Settings.Services {

		var latestVersionFormatted string
		if len(service.Version.Latest) == 0 || service.Version.Latest == "n/a" {
			latestVersionFormatted = "n/a"
		} else {
			if service.Version.Latest != "n/a" && version.IsVersionNewer(service.Version.Current, service.Version.Latest) {
				latestVersionFormatted = colorGreen + service.Version.Latest + colorReset
			} else {
				latestVersionFormatted = service.Version.Latest
			}
		}

		var devVersionFormatted string
		if len(service.Version.Dev) == 0 || service.Version.Dev == "n/a" {
			devVersionFormatted = "n/a"
		} else {
			devVersionFormatted = colorYellow + service.Version.Dev + colorReset
		}

		logger.LogInfo(fmt.Sprintf("%s %s (latest: %s, dev: %s)\n",
			service.BinName, service.Version.Current, latestVersionFormatted, devVersionFormatted))
	}
}
