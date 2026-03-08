package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "now",
	Short: "Keep your opennow.dev status green",
	Long:  "now auto-detects your context and pushes status updates to opennow.dev.",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
