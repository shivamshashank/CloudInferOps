package utils

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
)

// Standard diagnostic prefixes
const (
	PrefixOK    = ColorGreen + "  🟢  " + ColorReset
	PrefixWarn  = ColorYellow + "  🟡  " + ColorReset
	PrefixInfo  = ColorCyan + "  🔵  " + ColorReset
	PrefixError = ColorRed + "  🔴  " + ColorReset
	PrefixReady = ColorGreen + ColorBold + "  🚀  " + ColorReset
)

// ExecCommand runs a shell command and returns trimmed stdout, stderr, and any error.
func ExecCommand(dir string, name string, arg ...string) (string, string, error) {
	cmd := exec.Command(name, arg...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdout := strings.TrimSpace(stdoutBuf.String())
	stderr := strings.TrimSpace(stderrBuf.String())

	return stdout, stderr, err
}

// ExecCommandInteractive runs a command with stdin and stderr attached to the terminal while still capturing output.
func ExecCommandInteractive(dir string, name string, arg ...string) (string, string, error) {
	cmd := exec.Command(name, arg...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()
	stdout := strings.TrimSpace(stdoutBuf.String())
	stderr := strings.TrimSpace(stderrBuf.String())

	return stdout, stderr, err
}

// ExecCommandEnv runs a command with additional environment variables.
func ExecCommandEnv(dir string, env map[string]string, name string, arg ...string) (string, string, error) {
	cmd := exec.Command(name, arg...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = os.Environ()
	for key, value := range env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdout := strings.TrimSpace(stdoutBuf.String())
	stderr := strings.TrimSpace(stderrBuf.String())

	return stdout, stderr, err
}

// GetLocalIP returns the first non-loopback local IPv4 address of the active interfaces
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

// GetPublicIP robustly fetches the public IPv4 address of the machine.
// Iterates through multiple providers and cloud metadata endpoints to bypass rate limits or blocks.
func GetPublicIP() string {
	client := http.Client{Timeout: 3 * time.Second}
	providers := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
		"http://169.254.169.254/latest/meta-data/public-ipv4", // AWS EC2 / OpenStack metadata fallback
		"http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip",                    // GCP metadata fallback
		"http://169.254.169.254/metadata/instance/network/interface/0/ipv4/ipAddress/0/publicIpAddress?api-version=2021-02-01&format=text", // Azure metadata fallback
		"http://169.254.169.254/metadata/v1/interfaces/public/0/ipv4/address",                                                              // DigitalOcean metadata fallback
		"http://169.254.169.254/v1/interfaces/0/ipv4/address",                                                                              // Vultr metadata fallback
		"http://169.254.169.254/hetzner/v1/metadata/public-ipv4",                                                                           // Hetzner metadata fallback
	}

	for _, provider := range providers {
		req, err := http.NewRequest("GET", provider, nil)
		if err != nil {
			continue
		}
		// Masquerade as curl to bypass basic bot blockers on public IP APIs
		req.Header.Set("User-Agent", "curl/7.68.0")
		req.Header.Set("Metadata-Flavor", "Google") // Required by GCP metadata endpoint
		req.Header.Set("Metadata", "true")          // Required by Azure metadata endpoint

		resp, err := client.Do(req)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			ip := strings.TrimSpace(string(body))

			// Verify we received a valid public IP
			if parsedIP := net.ParseIP(ip); parsedIP != nil && !parsedIP.IsPrivate() && !parsedIP.IsLoopback() {
				return ip
			}
		}
	}
	return ""
}

// IsCloudVM attempts to determine if the Linux system is running inside a known cloud provider
// (AWS, GCP, Azure, Oracle, etc.) by reading DMI sys_vendor data.
func IsCloudVM() bool {
	data, err := os.ReadFile("/sys/class/dmi/id/sys_vendor")
	if err != nil {
		return false
	}
	vendor := strings.ToLower(strings.TrimSpace(string(data)))
	cloudVendors := []string{
		"amazon",
		"google",
		"microsoft",
		"oracle",
		"digitalocean",
		"linode",
		"vultr",
		"hetzner",
		"scaleway",
		"upcloud",
	}
	for _, v := range cloudVendors {
		if strings.Contains(vendor, v) {
			return true
		}
	}
	return false
}

// ExecCommandStream runs a command while streaming stdout and stderr to the terminal in real-time.
// It also captures and returns the full output for error reporting.
func ExecCommandStream(dir string, name string, arg ...string) (string, string, error) {
	cmd := exec.Command(name, arg...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()
	stdout := strings.TrimSpace(stdoutBuf.String())
	stderr := strings.TrimSpace(stderrBuf.String())

	return stdout, stderr, err
}

// GetRealHomeDir returns the home directory of the actual user,
// even when the command is run under sudo/root on Linux.
func GetRealHomeDir() (string, error) {
	if os.Getuid() == 0 {
		if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
			u, err := user.Lookup(sudoUser)
			if err == nil {
				return u.HomeDir, nil
			}
		}
	}
	return os.UserHomeDir()
}

// StartSpinner starts a background CLI spinner with the given message.
// It returns a function that must be called to stop the spinner and clear the line.
// It also listens for SIGINT/SIGTERM to gracefully handle Ctrl+C.
func StartSpinner(message string) func() {
	done := make(chan struct{})
	ack := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		// Print first frame immediately
		fmt.Printf("\r\033[K%s%s %s", PrefixInfo, message, frames[0])
		i++

		for {
			select {
			case <-done:
				fmt.Print("\r\033[K") // Clear line when done
				signal.Stop(sigChan)
				close(ack)
				return
			case <-sigChan:
				fmt.Print("\r\033[K") // Clear line on Ctrl+C
				fmt.Printf("%sOperation cancelled by user (Ctrl+C).\n", PrefixWarn)
				os.Exit(130) // Standard exit code for SIGINT
			case <-ticker.C:
				fmt.Printf("\r\033[K%s%s %s", PrefixInfo, message, frames[i%len(frames)])
				i++
			}
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			close(done)
			<-ack // Wait for goroutine to clear the line before returning
		})
	}
}
