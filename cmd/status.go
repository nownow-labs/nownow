package cmd

import (
	"fmt"

	"github.com/nownow-labs/nownow/internal/api"
	"github.com/nownow-labs/nownow/internal/config"
	"github.com/nownow-labs/nownow/internal/daemon"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show your current status on the board",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if !cfg.HasToken() {
			return fmt.Errorf("not logged in — run: nownow login")
		}

		// Daemon status
		if running, pid := daemon.IsRunning(); running {
			fmt.Printf("daemon: running (pid %d)\n", pid)
		} else {
			fmt.Println("daemon: not running")
		}

		client := api.NewClient(cfg.Endpoint, cfg.Token)

		// First get our identity
		me, err := client.VerifyToken()
		if err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}

		// Then get the board
		board, err := client.GetBoard()
		if err != nil {
			return fmt.Errorf("fetching board: %w", err)
		}

		// Find ourselves
		for _, entry := range board.Board {
			if entry.ID == me.User.ID {
				if entry.Status != "" {
					fmt.Printf("%s %s\n", entry.Emoji, entry.Status)
				} else {
					fmt.Println("(no status set)")
				}
				if entry.LastSeenAt != "" {
					fmt.Printf("last seen: %s\n", entry.LastSeenAt)
				}
				return nil
			}
		}

		fmt.Println("(not on the board yet)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
