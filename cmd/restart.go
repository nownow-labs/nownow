package cmd

import (
	"fmt"

	"github.com/opennow-labs/now-cli/internal/daemon"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the now daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		stopped, err := daemon.Stop()
		if err != nil {
			return err
		}
		if stopped {
			fmt.Println("daemon stopped")
		}
		return daemon.StartDaemon(true)
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
