package main

import (
	"fmt"

	"TDES/internals/drive"

	"github.com/spf13/cobra"
)

var driveRecipientPublicKey string

var driveCmd = &cobra.Command{
	Use:   "drive",
	Short: "Manage local drive-backed exercise delivery state",
	Long:  `Prepare and inspect local drive-backed storage used for exercise delivery and submissions.`,
}

var drivePrepareCmd = &cobra.Command{
	Use:   "prepare [drive-path]",
	Short: "Prepare a drive directory for exercise delivery",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := drive.PrepareDrive(args[0]); err != nil {
			fmt.Println("Error preparing drive:", err.Error())
			return
		}
		fmt.Println("Drive prepared at:", args[0])
	},
}

var drivePrepareSubmissionCmd = &cobra.Command{
	Use:   "prepare-submission [drive-path]",
	Short: "Prepare a drive directory for encrypted submissions",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := drive.PrepareDriveForSubmission(args[0], driveRecipientPublicKey); err != nil {
			fmt.Println("Error preparing drive submission module:", err.Error())
			return
		}
		fmt.Println("Drive submission module prepared at:", args[0])
	},
}

func init() {
	rootCmd.AddCommand(driveCmd)
	driveCmd.AddCommand(drivePrepareCmd)
	driveCmd.AddCommand(drivePrepareSubmissionCmd)
	drivePrepareSubmissionCmd.Flags().StringVar(
		&driveRecipientPublicKey,
		"recipient-public-key",
		"",
		"Base64-encoded X25519 recipient public key used to encrypt submissions",
	)
	_ = drivePrepareSubmissionCmd.MarkFlagRequired("recipient-public-key")
}
