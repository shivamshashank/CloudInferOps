package utils

import (
	"bytes"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
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
