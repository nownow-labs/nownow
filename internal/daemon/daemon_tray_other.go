//go:build !darwin

package daemon

import (
	"time"

	"github.com/nownow-labs/nownow/internal/tray"
)

func startTray(interval time.Duration) {
	tray.Run(interval)
}
