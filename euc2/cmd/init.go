package cmd

import (
	initMod "euc2/internals/init"
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize an exercise from the local cache",
	Long:  `Initialize an exercise by lab id and optional version into a working directory.`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		idWithVersion := args[0]
		workingDir := args[1]
		if idWithVersion == "" || workingDir == "" {
			fmt.Println("Error: idWithVersion and workingDir are required")
			return
		}
		id, version := initMod.SplitIdWithVersion(idWithVersion)
		err := initMod.InitFromID(id, version, workingDir)
		if err != nil {
			fmt.Println("Error initializing from ID:", err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
