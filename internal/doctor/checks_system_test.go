package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckOS(t *testing.T) {
	res := CheckOS()
	if res.Name != "OS Support" {
		t.Errorf("Expected CheckResult Name 'OS Support', got '%s'", res.Name)
	}
	// Just ensure it returns a valid status (StatusOK or StatusError)
	if res.Status != StatusOK && res.Status != StatusError {
		t.Errorf("Expected StatusOK or StatusError, got %v", res.Status)
	}
}

func TestCheckInternet(t *testing.T) {
	res := CheckInternet()
	if res.Name != "Internet Connection" {
		t.Errorf("Expected CheckResult Name 'Internet Connection', got '%s'", res.Name)
	}
}

func TestCheckCPU(t *testing.T) {
	res := CheckCPU()
	if res.Name != "CPU Cores" {
		t.Errorf("Expected CheckResult Name 'CPU Cores', got '%s'", res.Name)
	}
}

func TestCheckMemory(t *testing.T) {
	res := CheckMemory()
	if res.Name != "System Memory" {
		t.Errorf("Expected CheckResult Name 'System Memory', got '%s'", res.Name)
	}
}

func TestGetLinuxMemory(t *testing.T) {
	tmpDir := t.TempDir()
	meminfoFile := filepath.Join(tmpDir, "meminfo")
	content := "MemTotal:        8192000 kB\nMemFree:         4096000 kB\n"
	if err := os.WriteFile(meminfoFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write mock meminfo: %v", err)
	}

	oldPath := procMeminfoPath
	procMeminfoPath = meminfoFile
	defer func() { procMeminfoPath = oldPath }()

	mem, err := getLinuxMemory()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	expected := uint64(8192000 * 1024)
	if mem != expected {
		t.Errorf("expected memory %d bytes, got %d", expected, mem)
	}

	// 2. Failure: missing field
	badContent := "MemFree:         4096000 kB\n"
	if err := os.WriteFile(meminfoFile, []byte(badContent), 0644); err != nil {
		t.Fatalf("failed to write mock meminfo: %v", err)
	}
	_, err = getLinuxMemory()
	if err == nil {
		t.Error("expected error for missing MemTotal, got nil")
	}
}
