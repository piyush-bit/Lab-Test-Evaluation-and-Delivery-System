package main

import (
	"TDES/internals/drive"
	"TDES/internals/remote"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type submitStrategy struct {
	name      string
	shortHand string
	path      string
}

var submitStrategies = []submitStrategy{
	{
		name:      "drive",
		shortHand: "d",
	},
	{
		name:      "remote",
		shortHand: "r",
	},
}

var submitOrgID string
var submitStudentID string

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit the current exercise using exactly one strategy",
	Long:  `Submit the current exercise using exactly one destination strategy such as drive or remote.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		strategy, err := resolveSubmitStrategy(submitStrategies)
		if err != nil {
			cmd.Println("Error resolving submit strategy:", err.Error())
			return
		}

		submissionPath, err := submitExercise(strategy, submitOrgID, submitStudentID)
		if err != nil {
			cmd.Println("Error submitting to source:", strategy.name, ":", err.Error())
			return
		}

		cmd.Println("Submission result:", submissionPath)
	},
}

func submitExercise(strategy submitStrategy, orgID string, studentID string) (string, error) {
	switch strategy.name {
	case "drive":
		driveRef := &drive.Drive{Path: strategy.path}
		return driveRef.CreateSubmission(drive.SubmissionRequest{
			ExercisePath: ".",
			OrgID:        orgID,
			StudentID:    studentID,
		})
	case "remote":
		remoteRef := remote.NewRemote(strategy.path)
		return remoteRef.SubmitRemote(remote.SubmitRequest{
			ExercisePath: ".",
			OrgID:        orgID,
			StudentID:    studentID,
			BearerToken:  os.Getenv(remote.BearerTokenEnvVar),
		})
	default:
		return "", fmt.Errorf("unsupported submit strategy %q", strategy.name)
	}
}

func resolveSubmitStrategy(strategies []submitStrategy) (submitStrategy, error) {
	var active []submitStrategy

	for _, strategy := range strategies {
		if strings.TrimSpace(strategy.path) == "" {
			continue
		}
		active = append(active, strategy)
	}

	if len(active) == 0 {
		return submitStrategy{}, fmt.Errorf("exactly one strategy is required; pass one of --drive or --remote")
	}
	if len(active) > 1 {
		return submitStrategy{}, fmt.Errorf("only one submit strategy is permitted at a time")
	}

	return active[0], nil
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().StringVarP(
		&submitStrategies[0].path,
		submitStrategies[0].name,
		submitStrategies[0].shortHand,
		"",
		"Submit using the drive strategy and the provided destination path",
	)
	submitCmd.Flags().StringVarP(
		&submitStrategies[1].path,
		submitStrategies[1].name,
		submitStrategies[1].shortHand,
		"",
		"Submit using the remote strategy and the provided base URL",
	)
	submitCmd.Flags().StringVar(
		&submitOrgID,
		"org-id",
		"",
		"Organization identifier to embed in the submission package",
	)
	submitCmd.Flags().StringVar(
		&submitStudentID,
		"student-id",
		"",
		"Student identifier to embed in the submission package",
	)
}
