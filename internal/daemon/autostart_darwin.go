//go:build darwin

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"text/template"
)

const launchdLabel = "dev.opennow.cli"

var launchdPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>{{.Label}}</string>
  <key>ProgramArguments</key>
  <array>
    <string>{{.Exe}}</string>
    <string>start</string>
    <string>--foreground</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>{{.LogDir}}/now.log</string>
  <key>StandardErrorPath</key>
  <string>{{.LogDir}}/now.err</string>
</dict>
</plist>
`

func plistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", launchdLabel+".plist")
}

// IsAutostartInstalled returns true if the launchd plist exists.
func IsAutostartInstalled() bool {
	_, err := os.Stat(plistPath())
	return err == nil
}

// InstallAutostart creates a launchd plist for login startup.
func InstallAutostart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	logDir := filepath.Join(home, "Library", "Logs", "now")
	os.MkdirAll(logDir, 0700)

	tmpl, err := template.New("plist").Parse(launchdPlist)
	if err != nil {
		return err
	}

	f, err := os.Create(plistPath())
	if err != nil {
		return fmt.Errorf("creating plist: %w", err)
	}

	err = tmpl.Execute(f, map[string]string{
		"Label":  launchdLabel,
		"Exe":    exe,
		"LogDir": logDir,
	})
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}

	// Try to load the service into launchd so it takes effect immediately.
	// This is best-effort: bootstrap may fail in non-GUI sessions (SSH, CI)
	// or when the daemon is already running. The plist with RunAtLoad=true
	// guarantees it will load on next login regardless.
	if domain, err := guiDomain(); err == nil && !isServiceLoaded() {
		if err := exec.Command("launchctl", "bootstrap", domain, plistPath()).Run(); err != nil {
			fmt.Fprintf(os.Stderr, "note: launchctl bootstrap skipped (%v), will activate on next login\n", err)
		}
	}

	fmt.Printf("autostart installed: %s\n", plistPath())
	return nil
}

func guiDomain() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("getting current user: %w", err)
	}
	return "gui/" + u.Uid, nil
}

func isServiceLoaded() bool {
	domain, err := guiDomain()
	if err != nil {
		return false
	}
	return exec.Command("launchctl", "print", domain+"/"+launchdLabel).Run() == nil
}

// UninstallAutostart removes the launchd plist.
func UninstallAutostart() error {
	p := plistPath()
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil
	}
	// Unload the service from launchd before removing the plist.
	if isServiceLoaded() {
		domain, err := guiDomain()
		if err == nil {
			_ = exec.Command("launchctl", "bootout", domain+"/"+launchdLabel).Run()
		}
	}
	if err := os.Remove(p); err != nil {
		return err
	}
	fmt.Println("autostart removed")
	return nil
}
