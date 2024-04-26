package version

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"alex-hendel.de/gitea-update/internal/config"
	"alex-hendel.de/gitea-update/internal/logger"
	"github.com/Masterminds/semver/v3"
)

func CheckInstalledVersion(path, binaryName string) string {
	fullPath := fmt.Sprintf("%s/%s", path, binaryName)

	cmd := exec.Command(fullPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		logger.LogWarn(fmt.Sprintf("Error executing %s --version: %s\n", binaryName, err))
		return ""
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	if scanner.Scan() {
		line := scanner.Text()
		return extractVersion(line, binaryName)
	}

	if err := scanner.Err(); err != nil {
		logger.LogError(fmt.Sprintf("Error reading output from %s --version: %s\n", binaryName, err))
	}

	return ""
}

func extractVersion(line, binaryName string) string {
	switch binaryName {
	case "gitea":
		// Expected: "Gitea version 1.21.10 built with..."
		parts := strings.Split(line, " ")
		for i, part := range parts {
			if part == "version" && i+1 < len(parts) {
				return strings.TrimSpace(parts[i+1])
			}
		}
	case "act_runner":
		// Expected: "act_runner version v0.2.9"
		parts := strings.Split(line, " ")
		for i, part := range parts {
			if part == "version" && i+1 < len(parts) {
				return strings.TrimPrefix(strings.TrimSpace(parts[i+1]), "v")
			}
		}
	}
	return ""
}

func RequestVersion(service *config.Service) {
	apiURL := service.URLs.API

	resp, err := http.Get(apiURL)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error fetching version information for %s: %v", service.BinName, err))
		return
	}
	defer resp.Body.Close()

	var tags []struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		logger.LogError(fmt.Sprintf("Error decoding JSON response for %s: %v", service.BinName, err))
		return
	}

	var currentVersion *semver.Version
	if service.Version.Current != "n/a" {
		var err error
		currentVersion, err = semver.NewVersion(service.Version.Current)
		if err != nil {
			logger.LogError(fmt.Sprintf("Current version format is invalid for %s: %v", service.BinName, err))
			// Continue processing without a valid current version
		}
	}

	for _, tag := range tags {
		versionStr := strings.TrimPrefix(tag.Name, "v")
		semVer, err := semver.NewVersion(versionStr)
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid version format %s for %s: %v", versionStr, service.BinName, err))
			continue
		}

		// Update only if no valid current version or if the found version is greater
		if strings.Contains(tag.Name, "-dev") || strings.Contains(tag.Name, "-rc") {
			if service.Version.Dev == "n/a" || (currentVersion != nil && semVer.GreaterThan(currentVersion)) {
				service.Version.Dev = versionStr
				config.UpdateService(service.BinName, service)
			}
		} else {
			if service.Version.Latest == "n/a" || (currentVersion != nil && semVer.GreaterThan(currentVersion)) {
				service.Version.Latest = versionStr
				config.UpdateService(service.BinName, service)
			}
		}
	}
}

func isSemVer(version string) bool {
	_, err := semver.NewVersion(version)
	return err == nil
}

func IsVersionNewer(version1, version2 string) bool {
	if version1 == "n/a" && version2 == "n/a" {
		return false
	}

	if version1 == "n/a" && version2 != "n/a" {
		logger.LogDebug("Version 1 is n/a, checking if version 2 is a valid SemVer")
		return isSemVer(version2)
	}

	if version1 != "n/a" && version2 == "n/a" {
		return false
	}

	v1, err1 := semver.NewVersion(version1)
	v2, err2 := semver.NewVersion(version2)
	if err1 != nil || err2 != nil {
		logger.LogError(fmt.Sprintln("Error parsing versions:", err1, err2))
		return false
	}
	return v1.GreaterThan(v2)
}
