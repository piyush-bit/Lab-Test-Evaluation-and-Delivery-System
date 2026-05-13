package cmd

import (
	"fmt"
	"path/filepath"

	"euc2/internals/exercise"
	exercisestore "euc2/internals/exercise_store"

	"github.com/spf13/cobra"
)

var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Package an exercise and persist public and private artifacts locally",
	Long:  `Package an exercise directory into public and private artifacts, cache the public package, and save the private package into the local private store.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Println("Error converting path to absolute:", err.Error())
			return
		}
		publicPackages, privatePackages, err := exercise.PackageExercise(absPath)
		if err != nil {
			fmt.Println("Error running package function:", err.Error())
			return
		}
		if err := exercisestore.SavePackage(exercisestore.GetPublicCacheDir(), publicPackages); err != nil {
			fmt.Println("Error saving public package to cache:", err.Error())
			return
		}
		if err := exercisestore.SavePackage(exercisestore.GetPrivateCacheDir(), privatePackages); err != nil {
			fmt.Println("Error saving private package to local store:", err.Error())
			return
		}
		fmt.Println("Public packages:", publicPackages)
		fmt.Println("Private packages:", privatePackages)
	},
}

func init() {
	rootCmd.AddCommand(packageCmd)
}
