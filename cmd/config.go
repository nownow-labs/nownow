package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/opennow-labs/now-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Open config file in your editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure config exists
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return err
		}

		p, err := config.Path()
		if err != nil {
			return err
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			switch runtime.GOOS {
			case "darwin":
				editor = "open"
			case "windows":
				editor = "notepad"
			default:
				editor = "vi"
			}
		}

		fmt.Printf("opening %s\n", p)
		c := exec.Command(editor, p)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
