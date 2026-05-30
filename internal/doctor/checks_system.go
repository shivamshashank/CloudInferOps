package doctor

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// CheckOS validates OS and architecture compatibility
func CheckOS() CheckResult {
	supportedOS := map[string]bool{"linux": true}
	supportedArch := map[string]bool{"amd64": true, "arm64": true}

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	msg := fmt.Sprintf("OS: %s/%s", goos, goarch)

	if !supportedOS[goos] || !supportedArch[goarch] {
		return CheckResult{
			Name:   "OS Support",
			Status: StatusError,
			Message: msg + " (Unsupported OS/Arch. Only Linux and amd64/arm64 are supported)",
		}
	}

	return CheckResult{
		Name:   "OS Support",
		Status: StatusOK,
		Message: msg,
	}
}

// CheckInternet validates active network connection and DNS resolution
func CheckInternet() CheckResult {
	// 1. Direct connection test to Cloudflare DNS
	d := net.Dialer{Timeout: 3 * time.Second}
	conn, err := d.Dial("tcp", "1.1.1.1:53")
	if err != nil {
		return CheckResult{
			Name:   "Internet Connection",
			Status: StatusError,
			Message: "Internet connection (No internet access - failed to connect to 1.1.1.1:53)",
		}
	}
	conn.Close()

	// 2. DNS resolution test
	_, err = net.LookupHost("github.com")
	if err != nil {
		return CheckResult{
			Name:   "Internet Connection",
			Status: StatusError,
			Message: "Internet connection (DNS resolution failed - github.com unreachable)",
		}
	}

	return CheckResult{
		Name:   "Internet Connection",
		Status: StatusOK,
		Message: "Internet connection",
	}
}

// CheckCPU detects CPU cores and warns if below minimum requirement (2 cores)
func CheckCPU() CheckResult {
	cores := runtime.NumCPU()
	msg := fmt.Sprintf("Minimum CPU: %d cores (2 cores+ recommended)", cores)

	if cores < 2 {
		return CheckResult{
			Name:   "CPU Cores",
			Status: StatusWarn,
			Message: msg + " - CPU might be insufficient for full observability stack",
		}
	}

	return CheckResult{
		Name:   "CPU Cores",
		Status: StatusOK,
		Message: msg,
	}
}

// CheckMemory detects system memory and warns if below minimum requirement (4GB)
func CheckMemory() CheckResult {
	var totalBytes uint64
	var err error

	switch runtime.GOOS {
	case "linux":
		totalBytes, err = getLinuxMemory()
	default:
		err = fmt.Errorf("unsupported OS")
	}

	if err != nil {
		return CheckResult{
			Name:   "System Memory",
			Status: StatusWarn,
			Message: fmt.Sprintf("Minimum memory: Unknown (%v) - 4GB+ recommended", err),
		}
	}

	totalGB := float64(totalBytes) / (1024 * 1024 * 1024)
	msg := fmt.Sprintf("Minimum memory: %.2fGB (4GB+ recommended)", totalGB)

	if totalGB < 4.0 {
		return CheckResult{
			Name:   "System Memory",
			Status: StatusWarn,
			Message: msg + " - Observability components might suffer from OOM kills",
		}
	}

	return CheckResult{
		Name:   "System Memory",
		Status: StatusOK,
		Message: msg,
	}
}



func getLinuxMemory() (uint64, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return 0, err
				}
				// /proc/meminfo outputs MemTotal in kB, convert to bytes
				return val * 1024, nil
			}
		}
	}
	return 0, fmt.Errorf("MemTotal field not found in /proc/meminfo")
}
