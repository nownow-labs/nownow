//go:build !darwin

package daemon

import (
	"time"

	"github.com/opennow-labs/now-cli/internal/tray"
)

func startTray(interval time.Duration) {
	tray.Run(interval)
}
