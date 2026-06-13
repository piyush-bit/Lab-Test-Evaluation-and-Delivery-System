package main

import (
	"TDES/internals/remote"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	publishRemoteURL  string
	publishOrgID      string
	publishExerciseID string
	publishVersion    string
	publishStatus     string
)

var publishCmd = &cobra.Command{
	Use:   "publish [exercise-dir]",
	Short: "Package and publish an exercise to the remote registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exerciseDir := args[0]

		if publishRemoteURL == "" {
			publishRemoteURL = os.Getenv("EUC2_REGISTRY_URL")
		}
		if publishRemoteURL == "" {
			fmt.Println("Error: remote registry URL is required via --remote or EUC2_REGISTRY_URL")
			return
		}
		if publishOrgID == "" {
			fmt.Println("Error: organization ID is required via --org-id")
			return
		}

		remoteRef := remote.NewRemote(publishRemoteURL)
		response, err := remoteRef.PublishRemote(remote.PublishRequest{
			ExercisePath: exerciseDir,
			OrgID:        publishOrgID,
			ExerciseID:   publishExerciseID,
			Version:      publishVersion,
			Status:       publishStatus,
			BearerToken:  os.Getenv(remote.BearerTokenEnvVar),
		})
		if err != nil {
			fmt.Println("Error publishing exercise:", err.Error())
			return
		}

		fmt.Println("Publish result:", response)
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
	publishCmd.Flags().StringVarP(&publishRemoteURL, "remote", "r", "", "Remote registry base URL")
	publishCmd.Flags().StringVar(&publishOrgID, "org-id", "", "Organization ID")
	publishCmd.Flags().StringVar(&publishExerciseID, "exercise-id", "", "Optional override for exercise ID")
	publishCmd.Flags().StringVar(&publishVersion, "version", "", "Optional override for exercise version")
	publishCmd.Flags().StringVar(&publishStatus, "status", "", "Optional exercise status (default 'published')")
}
