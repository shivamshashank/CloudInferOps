package doctor

import (
	"runtime"
	"testing"
)

func TestCheckOS(t *testing.T) {
	result := CheckOS()
	// OS must be either OK or Error depending on platform, but since it is tested on macOS or Linux, it should be StatusOK.
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if result.Status != StatusOK {
			t.Errorf("expected StatusOK on %s, got %v with message: %s", runtime.GOOS, result.Status, result.Message)
		}
	} else {
		if result.Status != StatusError {
			t.Errorf("expected StatusError on unsupported OS %s, got %v", runtime.GOOS, result.Status)
		}
	}
}

func TestCheckCPU(t *testing.T) {
	result := CheckCPU()
	if result.Name != "CPU Cores" {
		t.Errorf("expected check name 'CPU Cores', got '%s'", result.Name)
	}
	
	cores := runtime.NumCPU()
	if cores >= 2 {
		if result.Status != StatusOK {
			t.Errorf("expected StatusOK for cores >= 2, got %v", result.Status)
		}
	} else {
		if result.Status != StatusWarn {
			t.Errorf("expected StatusWarn for cores < 2, got %v", result.Status)
		}
	}
}

func TestCheckMemory(t *testing.T) {
	result := CheckMemory()
	if result.Name != "System Memory" {
		t.Errorf("expected check name 'System Memory', got '%s'", result.Name)
	}

	// It should return either StatusOK or StatusWarn depending on the test machine's RAM size, but it should not return StatusError or panic.
	if result.Status != StatusOK && result.Status != StatusWarn {
		t.Errorf("expected StatusOK or StatusWarn for memory check, got status %v (message: %s)", result.Status, result.Message)
	}
}
