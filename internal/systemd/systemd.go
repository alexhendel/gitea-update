package systemd

import (
	"fmt"
	"os/exec"

	"alex-hendel.de/gitea-update/internal/logger"
)

func StopService(serviceName string) error {
	cmd := exec.Command("systemctl", "stop", serviceName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %v", serviceName, err)
	}
	logger.LogInfo(fmt.Sprintf("Service %s stopped successfully.\n", serviceName))
	return nil
}

func StartService(serviceName string) error {
	cmd := exec.Command("systemctl", "start", serviceName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to start service %s: %v", serviceName, err)
	}
	logger.LogInfo(fmt.Sprintf("Service %s started successfully.\n", serviceName))
	return nil
}

func RestartService(serviceName string) error {
	cmd := exec.Command("systemctl", "restart", serviceName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to restart service %s: %v", serviceName, err)
	}
	logger.LogInfo(fmt.Sprintf("Service %s restarted successfully.\n", serviceName))
	return nil
}
