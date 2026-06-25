package gitops

import (
	"fmt"
	"os/exec"

	"github.com/shivamshashank/CloudInferOps/internal/utils"
)

// CheckGitInstalled verifies that the git CLI is available on the host machine.
func CheckGitInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// InitLocalRepo initializes a git repo, configures local identity, adds all files, and commits.
func InitLocalRepo(repoDir string) error {
	// Initialize repository
	_, stderr, err := utils.ExecCommand(repoDir, "git", "init")
	if err != nil {
		return fmt.Errorf("failed to init git repo: %w (stderr: %s)", err, stderr)
	}

	// Configure local user
	_, stderr, err = utils.ExecCommand(repoDir, "git", "config", "user.name", "CloudInferOps Admin")
	if err != nil {
		return fmt.Errorf("failed to config user.name: %w (stderr: %s)", err, stderr)
	}

	_, stderr, err = utils.ExecCommand(repoDir, "git", "config", "user.email", "admin@cloudinfer.dev")
	if err != nil {
		return fmt.Errorf("failed to config user.email: %w (stderr: %s)", err, stderr)
	}

	// Track and commit all generated files
	_, stderr, err = utils.ExecCommand(repoDir, "git", "add", ".")
	if err != nil {
		return fmt.Errorf("failed to add files to git: %w (stderr: %s)", err, stderr)
	}

	_, _, err = utils.ExecCommand(repoDir, "git", "commit", "-m", "Bootstrap CloudInferOps GitOps")
	if err != nil {
		// If there is nothing to commit, we can ignore the error
		return nil
	}

	return nil
}

// PushToGitServer pushes the local repository to the in-cluster Git server.
func PushToGitServer(repoDir, remoteURL string) error {
	// Remove origin if it exists to allow update
	_, _, _ = utils.ExecCommand(repoDir, "git", "remote", "remove", "origin")

	// Add origin remote
	_, stderr, err := utils.ExecCommand(repoDir, "git", "remote", "add", "origin", remoteURL)
	if err != nil {
		return fmt.Errorf("failed to add remote: %w (stderr: %s)", err, stderr)
	}

	// Push to master/main (using force to override previous bootstrap deployments)
	_, _, err = utils.ExecCommand(repoDir, "git", "push", "-u", "origin", "master", "--force")
	if err != nil {
		// Try pushing to main if master fails
		_, stderr, err = utils.ExecCommand(repoDir, "git", "push", "-u", "origin", "main", "--force")
		if err != nil {
			return fmt.Errorf("failed to push to git server: %w (stderr: %s)", err, stderr)
		}
	}

	return nil
}
