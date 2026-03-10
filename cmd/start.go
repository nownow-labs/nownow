package cmd

import (
	"fmt"
	"time"

	"github.com/opennow-labs/now-cli/internal/config"
	"github.com/opennow-labs/now-cli/internal/daemon"
	"github.com/opennow-labs/now-cli/internal/tray"

	"github.com/spf13/cobra"
)

var (
	startInterval   string
	startForeground bool
	startNoAutostart bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the now daemon (background by default)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		intervalStr := startInterval
		if intervalStr == "" {
			intervalStr = cfg.Interval
		}
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return fmt.Errorf("invalid interval %q: %w", intervalStr, err)
		}

		if startForeground {
			// Run in foreground (used by detached process and launchd)
			if !cfg.HasToken() {
				return fmt.Errorf("not logged in — run: now login")
			}
			tray.Version = Version
			tray.RestartFunc = daemon.Restart
			return daemon.RunForeground(interval)
		}

		return daemon.StartDaemon(!startNoAutostart)
	},
}

func init() {
	startCmd.Flags().StringVar(&startInterval, "interval", "", "push interval (default from config, e.g. 5m)")
	startCmd.Flags().BoolVar(&startForeground, "foreground", false, "run in foreground (used internally)")
	startCmd.Flags().BoolVar(&startNoAutostart, "no-autostart", false, "skip autostart installation")
	startCmd.Flags().MarkHidden("foreground")
	rootCmd.AddCommand(startCmd)
}
