package install

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"

	"alex-hendel.de/gitea-update/internal/config"
	"alex-hendel.de/gitea-update/internal/logger"
	"alex-hendel.de/gitea-update/internal/systemd"
	"alex-hendel.de/gitea-update/internal/version"
)

func DownloadBinary(version, path, owner, group string, service *config.Service) error {
	downloadURL := service.URLs.Download
	downloadURL = strings.Replace(downloadURL, "{bin}", service.BinName, -1)
	downloadURL = strings.Replace(downloadURL, "{version}", version, -1)

	logger.LogDebug(fmt.Sprintf("Downloading %s version %s\n", service.BinName, version))
	logger.LogDebug(fmt.Sprintf("Downloading %s from %s\n", service.BinName, downloadURL))
	logger.LogDebug(fmt.Sprintf("Downloading to %s\n", path))

	response, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil {
		return err
	}

	err = os.Chmod(path, 0755)
	if err != nil {
		return fmt.Errorf("failed to set execute permissions: %v", err)
	}

	uid, gid, err := getUserGroupIds(owner, group)
	if err != nil {
		return fmt.Errorf("failed to get user/group IDs: %v", err)
	}

	err = os.Chown(path, uid, gid)
	if err != nil {
		return fmt.Errorf("failed to change file owner: %v", err)
	}

	var stderr bytes.Buffer
	var cmd = exec.Command("setcap", "cap_net_bind_service=+ep", path)
	cmd.Stderr = &stderr

	logger.LogDebug(cmd.String())
	err = cmd.Run()
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to set capabilities: %v, stderr: %s", err, stderr.String()))
		return fmt.Errorf("failed to set capabilities: %v, stderr: %s", err, stderr.String())
	}
	logger.LogDebug("Set capabilities successfully.")

	return nil
}

func getUserGroupIds(owner, group string) (int, int, error) {
	u, err := user.Lookup(owner)
	if err != nil {
		return -1, -1, fmt.Errorf("lookup user %s: %v", owner, err)
	}

	g, err := user.LookupGroup(group)
	if err != nil {
		return -1, -1, fmt.Errorf("lookup group %s: %v", group, err)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return -1, -1, fmt.Errorf("convert user id: %v", err)
	}

	gid, err := strconv.Atoi(g.Gid)
	if err != nil {
		return -1, -1, fmt.Errorf("convert group id: %v", err)
	}

	return uid, gid, nil
}

func PerformInstallation(dev bool) {
	for _, serviceName := range config.ListServices() {
		service, exists := config.GetService(serviceName)
		if !exists {
			logger.LogError(fmt.Sprintf("Service %s not found in configuration\n", serviceName))
			os.Exit(1)
		}

		targetVersion := service.Version.Latest
		if dev {
			targetVersion = service.Version.Dev
		}

		if targetVersion == "n/a" || len(targetVersion) == 0 {
			logger.LogError(fmt.Sprintf("Target version '%s' invalid for %s\n", targetVersion, service.BinName))
			os.Exit(1)
		}

		if err := systemd.StopService(service.SystemdName); err != nil {
			logger.LogError(fmt.Sprintf("Stopping service %s failed: %s\n", service.SystemdName, err))
			os.Exit(1)
		}

		currentBinaryPath := fmt.Sprintf("%s/%s", service.Path, service.BinName)
		backupBinaryPath := fmt.Sprintf("%s/%s.old", service.Path, service.BinName)

		if _, err := os.Stat(currentBinaryPath); err == nil {
			if err := os.Rename(currentBinaryPath, backupBinaryPath); err != nil {
				logger.LogWarn(fmt.Sprintf("Failed to backup the binary %s: %s\n", service.BinName, err))
				continue
			}
		} else if !os.IsNotExist(err) {
			logger.LogWarn(fmt.Sprintf("Error checking for existing binary %s: %s\n", service.BinName, err))
			continue
		}

		if err := DownloadBinary(targetVersion, currentBinaryPath, config.AppConfig.Settings.User, config.AppConfig.Settings.Group, &service); err != nil {
			logger.LogError(fmt.Sprintf("Failed to install new version of %s: %s\n", service.BinName, err))
			// Attempt to restore from backup if installation fails
			os.Rename(backupBinaryPath, currentBinaryPath)
			continue
		}

		if err := systemd.StartService(service.SystemdName); err != nil {
			logger.LogError(fmt.Sprintf("Starting service %s failed: %s\n", service.SystemdName, err))
			os.Exit(1)
		}

		logger.LogInfo(fmt.Sprintf("Successfully installed %s version %s\n", service.SystemdName, targetVersion))
	}
}

func RetrieveRemoteVersion() {
	var wg sync.WaitGroup

	for binary, service := range config.AppConfig.Settings.Services {
		wg.Add(1)
		go func(binary string, service config.Service) {
			defer wg.Done()
			serviceCopy := service

			version.RequestVersion(&serviceCopy)
			config.UpdateService(binary, &serviceCopy)
		}(binary, service)
	}

	wg.Wait()
}
