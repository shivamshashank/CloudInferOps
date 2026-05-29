package utils

import (
	"bytes"
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
