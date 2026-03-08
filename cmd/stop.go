package cmd

import (
	"fmt"

	"github.com/opennow-labs/now-cli/internal/daemon"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the now daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := daemon.Stop(); err != nil {
			return err
		}
		fmt.Println("daemon stopped")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
