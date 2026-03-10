package daemon

import (
	"fmt"
	"os"
	"os/exec"

	"fyne.io/systray"
)

// Restart replaces the current daemon with a new one.
// On macOS with launchd, it updates the plist and spawns a detached helper
// to bootout+bootstrap the service. Otherwise it spawns a new process directly.
func Restart() error {
	if IsServiceLoaded() {
		// Update the plist in case the binary path changed (e.g. after upgrade).
		// InstallAutostart skips bootstrap when service is already loaded.
		if err := InstallAutostart(); err != nil {
			return fmt.Errorf("updating plist: %w", err)
		}
		// Spawn a detached subprocess to do bootout+bootstrap.
		// We can't do it in-process because bootout sends SIGTERM to us.
		if err := launchdRestart(); err != nil {
			return fmt.Errorf("launchd restart: %w", err)
		}
		// Current process will be killed by launchd bootout's SIGTERM.
		return nil
	}

	// Fallback: manually spawn a new process (non-launchd / non-macOS)
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating executable: %w", err)
	}

	cmd := exec.Command(exe, "start", "--foreground")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = detachedProcAttr()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting new daemon: %w", err)
	}

	// Quit current systray (triggers cleanup via onExit)
	systray.Quit()
	return nil
}
