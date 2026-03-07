package daemon

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestPidFileLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Initially not running
	running, _ := IsRunning()
	if running {
		t.Error("should not be running initially")
	}

	// Write our own PID
	if err := WritePid(); err != nil {
		t.Fatalf("WritePid: %v", err)
	}

	// Should detect as running (our own process)
	running, pid := IsRunning()
	if !running {
		t.Error("should be running after WritePid")
	}
	if pid != os.Getpid() {
		t.Errorf("pid = %d, want %d", pid, os.Getpid())
	}

	// Verify file contents
	pidPath := filepath.Join(tmpDir, "nownow", "daemon.pid")
	data, err := os.ReadFile(pidPath)
	if err != nil {
		t.Fatalf("reading pid file: %v", err)
	}
	if string(data) != strconv.Itoa(os.Getpid()) {
		t.Errorf("pid file content = %q, want %q", string(data), strconv.Itoa(os.Getpid()))
	}

	// Remove PID
	RemovePid()
	running, _ = IsRunning()
	if running {
		t.Error("should not be running after RemovePid")
	}
}

func TestIsRunningWithDeadPid(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Write a PID that definitely doesn't exist
	pidDir := filepath.Join(tmpDir, "nownow")
	os.MkdirAll(pidDir, 0700)
	os.WriteFile(filepath.Join(pidDir, "daemon.pid"), []byte("999999999"), 0600)

	running, _ := IsRunning()
	if running {
		t.Error("should not detect dead PID as running")
	}
}

func TestIsRunningNoPidFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	running, _ := IsRunning()
	if running {
		t.Error("should not be running without pid file")
	}
}

func TestStopWhenNotRunning(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	err := Stop()
	if err == nil {
		t.Error("Stop should error when not running")
	}
}
