package daemon

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/opennow-labs/now-cli/internal/config"
	"github.com/opennow-labs/now-cli/internal/logging"
	"github.com/opennow-labs/now-cli/internal/settings"
	"github.com/opennow-labs/now-cli/internal/tray"
)

// PidFile returns the path to the daemon PID file.
func PidFile() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "daemon.pid"), nil
}

// IsRunning checks if a daemon process is alive.
func IsRunning() (bool, int) {
	pidPath, err := PidFile()
	if err != nil {
		return false, 0
	}
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return false, 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false, 0
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, 0
	}
	// Signal 0 checks if process exists
	if err := process.Signal(syscall.Signal(0)); err != nil {
		os.Remove(pidPath)
		return false, 0
	}
	return true, pid
}

// WritePid writes the current process PID to the pid file.
func WritePid() error {
	pidPath, err := PidFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(pidPath), 0700); err != nil {
		return err
	}
	return os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0600)
}

// RemovePid removes the pid file only if it still belongs to the current process.
// This prevents a restarting process from deleting a new process's pid file.
func RemovePid() {
	p, err := PidFile()
	if err != nil {
		return
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		os.Remove(p)
		return
	}
	if pid == os.Getpid() {
		os.Remove(p)
	}
}

// StartDaemon starts the daemon via the platform service manager if available,
// otherwise falls back to StartDetached. If installAutostart is true and no
// service manager is active, it also installs autostart for future logins.
func StartDaemon(installAutostart bool) error {
	// Fail fast if not logged in — avoids spawning a process that dies immediately.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if !cfg.HasToken() {
		return fmt.Errorf("not logged in — run: now login")
	}

	// Try platform service manager first (launchd on macOS, no-op elsewhere)
	managed, err := startViaServiceManager()
	if err != nil {
		return err
	}
	if managed {
		return nil
	}

	// No service manager — start manually
	if err := StartDetached(); err != nil {
		return err
	}
	if installAutostart && !IsAutostartInstalled() {
		if err := InstallAutostart(); err != nil {
			fmt.Printf("note: autostart setup skipped (%s)\n", err)
		}
	}
	return nil
}

// Stop stops the running daemon. On macOS, it uses launchd bootout when the
// service is managed by launchd, which prevents KeepAlive from restarting it.
// Falls back to SIGTERM for manually started processes.
// Returns (true, nil) if a daemon was actually stopped, (false, nil) if
// nothing was running, or (false, err) on failure.
func Stop() (stopped bool, err error) {
	running, pid := IsRunning()

	// Try service manager first — it may be managing the process even if
	// the PID file is missing (e.g. crash during throttled restart).
	managed, err := stopViaServiceManager()
	if err != nil {
		return false, err
	}

	if !managed && !running {
		return false, nil
	}

	if managed {
		// Bootout succeeded; wait for the process to exit if we can find it.
		if running {
			return true, waitForExit(pid)
		}
		fmt.Println("service unloaded from launchd")
		return true, nil
	}

	// Fallback: send SIGTERM directly (manually started / non-macOS)
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return false, fmt.Errorf("failed to stop daemon (pid %d): %w", pid, err)
	}
	return true, waitForExit(pid)
}

// waitForExit polls until the process with the given PID exits (up to 5 seconds).
func waitForExit(pid int) error {
	process, _ := os.FindProcess(pid) // always succeeds on Unix
	for i := 0; i < 50; i++ {
		if err := process.Signal(syscall.Signal(0)); err != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("daemon (pid %d) did not exit within 5 seconds", pid)
}

// StartDetached launches the daemon as a background process.
func StartDetached() error {
	if running, pid := IsRunning(); running {
		return fmt.Errorf("daemon already running (pid %d)", pid)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable: %w", err)
	}

	cmd := exec.Command(exe, "start", "--foreground")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	// Detach from parent process group
	cmd.SysProcAttr = detachedProcAttr()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	fmt.Printf("daemon started (pid %d)\n", cmd.Process.Pid)
	return nil
}

// RunForeground runs the menubar tray + push loop (called by detached process).
func RunForeground(interval time.Duration) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if !cfg.HasToken() {
		return fmt.Errorf("not logged in — run: now login")
	}

	// Initialize structured logging
	dir, err := config.Dir()
	if err != nil {
		return fmt.Errorf("config dir: %w", err)
	}
	if err := logging.Init(dir); err != nil {
		return fmt.Errorf("init logging: %w", err)
	}
	slog.Info("daemon starting", "version", tray.Version, "interval", interval, "pid", os.Getpid())

	if err := WritePid(); err != nil {
		return fmt.Errorf("writing pid: %w", err)
	}
	defer RemovePid()
	defer slog.Info("daemon exiting", "pid", os.Getpid())

	// Start settings HTTP server
	settings.AutostartIsInstalled = IsAutostartInstalled
	settings.AutostartInstall = InstallAutostart
	settings.AutostartUninstall = UninstallAutostart
	if err := settings.Start(tray.Version); err != nil {
		slog.Warn("settings UI unavailable", "error", err)
	} else {
		tray.SettingsAvailable = true
	}

	// Launch systray + settings window — this blocks on the main thread.
	// Platform-specific: on macOS uses a native webview window for settings,
	// on other platforms falls back to systray.Run with browser-based settings.
	startTray(interval)
	return nil
}

// InstallAutostart and UninstallAutostart are implemented per-platform
// in autostart_darwin.go, autostart_linux.go, autostart_windows.go.
