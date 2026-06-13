package main

import (
	"crypto/ecdh"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"TDES/internals/drive"

	"github.com/spf13/cobra"
)

var driveRecipientPublicKey string
var driveDecryptSubmissionRecipientPrivateKey string
var driveDecryptSubmissionOutput string

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

var driveDecryptSubmissionCmd = &cobra.Command{
	Use:   "decrypt-submission [envelope-json-path]",
	Short: "Decrypt an encrypted JSON submission envelope",
	Long:  `Decrypt an encrypted JSON submission envelope back into a plaintext tar archive using the recipient private key.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		envelopePath := args[0]

		envelopeData, err := os.ReadFile(envelopePath)
		if err != nil {
			fmt.Println("Error reading submission envelope:", err.Error())
			return
		}

		var envelope drive.SubmissionEnvelope
		if err := json.Unmarshal(envelopeData, &envelope); err != nil {
			fmt.Println("Error decoding submission envelope JSON:", err.Error())
			return
		}

		rawKey, err := base64.StdEncoding.DecodeString(strings.TrimSpace(driveDecryptSubmissionRecipientPrivateKey))
		if err != nil {
			fmt.Println("Error decoding recipient private key from base64:", err.Error())
			return
		}

		privateKey, err := ecdh.X25519().NewPrivateKey(rawKey)
		if err != nil {
			fmt.Println("Error parsing X25519 recipient private key:", err.Error())
			return
		}

		plaintext, err := drive.DecryptSubmissionArchive(envelope, privateKey)
		if err != nil {
			fmt.Println("Error decrypting submission archive:", err.Error())
			return
		}

		if err := os.WriteFile(driveDecryptSubmissionOutput, plaintext, 0644); err != nil {
			fmt.Println("Error writing decrypted archive:", err.Error())
			return
		}

		fmt.Println("Decrypted submission package written to:", driveDecryptSubmissionOutput)
	},
}

func init() {
	rootCmd.AddCommand(driveCmd)
	driveCmd.AddCommand(drivePrepareCmd)
	driveCmd.AddCommand(drivePrepareSubmissionCmd)
	driveCmd.AddCommand(driveDecryptSubmissionCmd)

	drivePrepareSubmissionCmd.Flags().StringVar(
		&driveRecipientPublicKey,
		"recipient-public-key",
		"",
		"Base64-encoded X25519 recipient public key used to encrypt submissions",
	)
	_ = drivePrepareSubmissionCmd.MarkFlagRequired("recipient-public-key")

	driveDecryptSubmissionCmd.Flags().StringVar(
		&driveDecryptSubmissionRecipientPrivateKey,
		"recipient-private-key",
		"",
		"Base64-encoded X25519 recipient private key used to decrypt the submission",
	)
	_ = driveDecryptSubmissionCmd.MarkFlagRequired("recipient-private-key")

	driveDecryptSubmissionCmd.Flags().StringVarP(
		&driveDecryptSubmissionOutput,
		"output",
		"o",
		"",
		"Path where the decrypted .tar file should be written",
	)
	_ = driveDecryptSubmissionCmd.MarkFlagRequired("output")
}
