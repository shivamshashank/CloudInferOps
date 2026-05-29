package helm

import (
	"fmt"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

// AddRepo registers a new Helm repository registry if not dry-run
func AddRepo(name, url string, dryRun bool) error {
	cmdArgs := []string{"repo", "add", name, url}
	if dryRun {
		fmt.Printf("%s[DRY-RUN] helm %s\n", utils.PrefixInfo, strings.Join(cmdArgs, " "))
		return nil
	}

	fmt.Printf("%sAdding Helm repository '%s' from %s...\n", utils.PrefixInfo, name, url)
	_, stderr, err := utils.ExecCommand("", "helm", cmdArgs...)
	if err != nil {
		return fmt.Errorf("failed to add helm repo '%s': %w (stderr: %s)", name, err, stderr)
	}

	fmt.Printf("%sSuccessfully registered Helm repository '%s'.\n", utils.PrefixOK, name)
	return nil
}

// UpdateRepos updates Helm chart registries if not dry-run
func UpdateRepos(dryRun bool) error {
	cmdArgs := []string{"repo", "update"}
	if dryRun {
		fmt.Printf("%s[DRY-RUN] helm %s\n", utils.PrefixInfo, strings.Join(cmdArgs, " "))
		return nil
	}

	fmt.Printf("%sUpdating Helm repositories...\n", utils.PrefixInfo)
	_, stderr, err := utils.ExecCommand("", "helm", cmdArgs...)
	if err != nil {
		return fmt.Errorf("failed to update helm registries: %w (stderr: %s)", err, stderr)
	}

	fmt.Printf("%sHelm repositories updated successfully.\n", utils.PrefixOK)
	return nil
}

// InstallRelease installs or upgrades a Helm chart release in the target namespace if not dry-run
func InstallRelease(release, chart, namespace string, flags []string, dryRun bool) error {
	cmdArgs := []string{"upgrade", "--install", release, chart, "-n", namespace, "--create-namespace"}
	if len(flags) > 0 {
		cmdArgs = append(cmdArgs, flags...)
	}

	if dryRun {
		fmt.Printf("%s[DRY-RUN] helm %s\n", utils.PrefixInfo, strings.Join(cmdArgs, " "))
		return nil
	}

	fmt.Printf("%sDeploying Helm release '%s' (%s) into namespace '%s'...\n", utils.PrefixInfo, release, chart, namespace)
	_, stderr, err := utils.ExecCommand("", "helm", cmdArgs...)
	if err != nil {
		return fmt.Errorf("failed to install helm release '%s': %w (stderr: %s)", release, err, stderr)
	}

	fmt.Printf("%sSuccessfully deployed Helm release '%s'.\n", utils.PrefixOK, release)
	return nil
}
