//go:build darwin

package daemon

import (
	"time"

	"fyne.io/systray"
	"github.com/opennow-labs/now-cli/internal/settings"
	"github.com/opennow-labs/now-cli/internal/tray"
	"github.com/opennow-labs/now-cli/internal/webwin"
)

func startTray(interval time.Duration) {
	// Wire up native window for Settings and Quit before starting systray,
	// so callbacks are ready when the menu is built.
	tray.ShowSettings = webwin.Show
	tray.QuitFunc = webwin.Terminate

	// Set up systray without starting the event loop.
	start, end := systray.RunWithExternalLoop(
		func() { tray.OnReady(interval) },
		tray.OnExit,
	)
	start()

	// Create hidden webview window for settings UI.
	webwin.Init("http://"+settings.ListenAddr, end)

	// Start the macOS event loop — this blocks and drives both
	// the systray and the webview until Terminate is called.
	webwin.RunEventLoop()
}
