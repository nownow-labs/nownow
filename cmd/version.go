package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Set via ldflags at build time: -ldflags "-X github.com/opennow-labs/now-cli/cmd.Version=v0.1.0"
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print now version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("now %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
