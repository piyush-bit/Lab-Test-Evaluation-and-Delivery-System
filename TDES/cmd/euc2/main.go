package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "euc2",
	Short: "CLI to package, cache, initialize, and run exercises",
	Long:  `euc2 is a CLI for packaging, caching, initializing, and running coding exercises.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	Execute()
}
