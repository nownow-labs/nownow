package main

import (
	"fmt"
	"os"

	"github.com/nownow-labs/nownow/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
