package cli

import (
	"fmt"
	"os"
	"os/user"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "cloudinferops",
	Short: "🚀 CloudInferOps is a cloud-native AI Inference Operations Platform for Kubernetes",
	Long: `🚀 CloudInferOps is an end-to-end AI Inference Operations Platform designed to simplify
the deployment, management, and observability of LLM inference workloads on Kubernetes.

It provides a powerful CLI to deploy the core platform, inference services, and more.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	setupKubeconfigEnv()

	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func setupKubeconfigEnv() {
	// Respect any existing KUBECONFIG environment variable
	if os.Getenv("KUBECONFIG") != "" {
		return
	}

	// 1. Try to load config.yaml and use configured kubeconfig path
	if err := config.InitConfig(false); err == nil {
		if config.GlobalConfig.Kubernetes.Kubeconfig != "" {
			_ = os.Setenv("KUBECONFIG", config.GlobalConfig.Kubernetes.Kubeconfig)
			return
		}
	}

	// 2. Fall back to standard user-specific kubeconfig
	defaultKubeconfig := config.ExpandPath("~/.kube/config")
	if _, err := os.Stat(defaultKubeconfig); err == nil {
		_ = os.Setenv("KUBECONFIG", defaultKubeconfig)
		return
	}

	// 3. Fall back to /etc/kubernetes/admin.conf if running as root/sudo
	if os.Getuid() == 0 || hasRootPrivileges() {
		adminConfig := "/etc/kubernetes/admin.conf"
		if _, err := os.Stat(adminConfig); err == nil {
			_ = os.Setenv("KUBECONFIG", adminConfig)
			return
		}
	}
}

func hasRootPrivileges() bool {
	return hasRootPrivilegesForTest(os.Geteuid, os.Getuid, os.Getenv, func() (*user.User, error) {
		return user.Current()
	})
}

func hasRootPrivilegesForTest(euid func() int, uid func() int, getenv func(string) string, currentUser func() (*user.User, error)) bool {
	if euid() == 0 || uid() == 0 {
		return true
	}

	if getenv("SUDO_USER") != "" || getenv("SUDO_UID") != "" || getenv("USER") == "root" {
		return true
	}

	if currentUser != nil {
		if usr, err := currentUser(); err == nil && usr != nil {
			if usr.Uid == "0" || usr.Gid == "0" {
				return true
			}
		}
	}

	return false
}
