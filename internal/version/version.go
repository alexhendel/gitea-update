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
		return "n/a"
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	if scanner.Scan() {
		line := scanner.Text()
		return extractVersion(line)
	}

	if err := scanner.Err(); err != nil {
		logger.LogError(fmt.Sprintf("Error reading output from %s --version: %s\n", binaryName, err))
	}

	return "n/a"
}

func extractVersion(line string) string {
	// Expected: "Gitea version 1.21.10 built with..." or "act_runner version v0.2.9"
	parts := strings.Split(line, " ")
	for i, part := range parts {
		if part == "version" && i+1 < len(parts) {
			return strings.TrimPrefix(strings.TrimSpace(parts[i+1]), "v")
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

	for _, tag := range tags {
		tagVersionStr := strings.TrimPrefix(tag.Name, "v")
		tagVersion, err := semver.NewVersion(tagVersionStr)
		if err != nil {
			logger.LogWarn(fmt.Sprintf("Tag contained invalid version format %s for %s: %v", tagVersionStr, service.BinName, err))
			continue
		}

		logger.LogDebug(fmt.Sprintf("Checking remote version tag %s for %s", tagVersionStr, service.BinName))

		isDevVersion := false
		if strings.Contains(tag.Name, "-dev") || strings.Contains(tag.Name, "-rc") {
			isDevVersion = true
		}

		currentVersionStr := "n/a"
		if isDevVersion {
			currentVersionStr = service.Version.Dev
		} else {
			currentVersionStr = service.Version.Latest
		}

		currentVersion, err := semver.NewVersion(currentVersionStr)
		if err != nil {
			if isDevVersion {
				service.Version.Dev = tagVersionStr
			} else {
				service.Version.Latest = tagVersionStr
			}
		} else {
			if tagVersion.GreaterThan(currentVersion) {
				if isDevVersion {
					service.Version.Dev = tagVersionStr
				} else {
					service.Version.Latest = tagVersionStr
				}
			}
		}
		config.UpdateService(service.BinName, service)
	}
}

func isSemVer(version string) *semver.Version {
	newVersion, err := semver.NewVersion(version)
	if err != nil {
		return nil
	}

	return newVersion
}

func IsVersionNewer(version1, version2 string) bool {
	if version1 == "n/a" && version2 == "n/a" {
		return false
	}

	if version1 == "n/a" && version2 != "n/a" {
		logger.LogDebug("Version 1 is n/a, checking if version 2 is a valid SemVer")

		return isSemVer(version2) != nil
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
	return v2.GreaterThan(v1)
}
