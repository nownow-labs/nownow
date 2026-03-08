//go:build darwin

package webwin

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa

#include <stdint.h>
#include <stdlib.h>

void webwin_hideWindow(uintptr_t wndPtr);
void webwin_showWindow(uintptr_t wndPtr);
void webwin_setWindowDelegate(uintptr_t wndPtr);
void webwin_setAppIcon(const void *data, int length);
*/
import "C"

import (
	"log/slog"

	"github.com/opennow-labs/now-cli/internal/webview"
)

var (
	onQuit func()
	wv     webview.WebView
)

// Init creates a hidden webview window pointing at the settings URL.
func Init(url string, quitFn func()) {
	onQuit = quitFn

	wv = webview.New(false)
	if wv == nil {
		slog.Error("failed to create webview window")
		return
	}
	wv.SetTitle("Now Settings")
	wv.SetSize(900, 700, webview.HintNone)
	wv.Navigate(url)

	// Hide window initially
	C.webwin_hideWindow(C.uintptr_t(uintptr(wv.Window())))
	// Set delegate so close button hides instead of destroying
	C.webwin_setWindowDelegate(C.uintptr_t(uintptr(wv.Window())))
}

// Show displays the settings window and brings it to front.
func Show() {
	if wv == nil {
		return
	}
	C.webwin_showWindow(C.uintptr_t(uintptr(wv.Window())))
}

// RunEventLoop starts the macOS event loop. This blocks until Terminate is called.
// On macOS, [NSApp run] drives both the systray and the webview.
func RunEventLoop() {
	if wv == nil {
		slog.Error("webview not initialized, cannot run event loop")
		if onQuit != nil {
			onQuit()
		}
		return
	}
	wv.Run()
	if onQuit != nil {
		onQuit()
	}
}

// Terminate stops the event loop, causing RunEventLoop to return.
func Terminate() {
	if wv != nil {
		wv.Terminate()
	}
}

// SetAppIcon sets the application icon from the embedded PNG.
// This makes the icon visible in Activity Monitor, Spotlight, etc.
func SetAppIcon() {
	if len(appIconPNG) == 0 {
		return
	}
	ptr := C.CBytes(appIconPNG)
	C.webwin_setAppIcon(ptr, C.int(len(appIconPNG)))
	C.free(ptr)
}
