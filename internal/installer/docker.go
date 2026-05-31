package installer

import (
	"fmt"
	"os"
	"time"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

// InstallDocker downloads and installs Docker Engine using the official convenience script.
func InstallDocker() error {
	fmt.Printf("%sDownloading and installing Docker Engine...\n", utils.PrefixInfo)

	var err error
	var stderr string
	if os.Getuid() == 0 {
		_, stderr, err = utils.ExecCommandStream("", "sh", "-c", "curl -fsSL https://get.docker.com | sh")
	} else {
		fmt.Printf("%sAdministrative privileges required. Requesting sudo password...\n", utils.PrefixInfo)
		_, stderr, err = utils.ExecCommandStream("", "sudo", "sh", "-c", "curl -fsSL https://get.docker.com | sh")
	}

	if err != nil {
		return fmt.Errorf("failed to install Docker: %w (stderr: %s)", err, stderr)
	}
	fmt.Printf("%sDocker Engine installed successfully.\n", utils.PrefixOK)

	// Ensure docker service is running
	fmt.Printf("%sStarting Docker service...\n", utils.PrefixInfo)
	if os.Getuid() == 0 {
		_, _, _ = utils.ExecCommand("", "systemctl", "enable", "--now", "docker")
	} else {
		_, _, _ = utils.ExecCommand("", "sudo", "systemctl", "enable", "--now", "docker")
	}

	// Wait for Docker daemon to be ready
	stopSpinner := utils.StartSpinner("Waiting for Docker daemon to initialize...")
	for i := 0; i < 15; i++ {
		_, _, err := utils.ExecCommand("", "docker", "info")
		if err == nil {
			stopSpinner()
			fmt.Printf("%sDocker daemon is ready.\n", utils.PrefixOK)
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	stopSpinner()

	return fmt.Errorf("docker installed but daemon failed to start in time")
}
