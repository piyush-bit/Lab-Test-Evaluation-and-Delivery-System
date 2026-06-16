package main

import (
	"crypto/ecdh"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"TDES/internals/drive"
	evaluatorcore "TDES/internals/evaluator-core"

	"github.com/spf13/cobra"
)

var (
	batchRecipientPrivateKey string
	batchPrivateStore        string
	batchDockerBinary        string
	batchRegistryURL         string
	batchBearerToken         string
	batchCSVPath             string
	batchJSONPath            string
	batchOutputDir           string
)

type BatchResultRecord struct {
	EnvelopePath string `json:"envelope_path"`
	StudentID    string `json:"student_id,omitempty"`
	LabID        string `json:"lab_id,omitempty"`
	Version      string `json:"version,omitempty"`
	Status       string `json:"status"`
	EarnedPoints int    `json:"earned_points"`
	MaxPoints    int    `json:"max_points"`
	Error        string `json:"error,omitempty"`
}

var driveEvaluateBatchCmd = &cobra.Command{
	Use:   "evaluate-batch [drive-path]",
	Short: "Decrypt and evaluate all student submissions in a local drive in batch",
	Long:  `Scan the submissions directory under the drive root, decrypt each envelope using the private key, run the grading sandbox, and generate summary reports.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		drivePath := args[0]

		// Decode the private key
		rawKey, err := base64.StdEncoding.DecodeString(strings.TrimSpace(batchRecipientPrivateKey))
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "Error decoding recipient private key from base64:", err.Error())
			return
		}

		privateKey, err := ecdh.X25519().NewPrivateKey(rawKey)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "Error parsing X25519 recipient private key:", err.Error())
			return
		}

		fmt.Println("Scanning and decrypting submissions in:", drivePath)
		decrypted, err := drive.LoadAndDecryptSubmissions(drivePath, privateKey)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "Error loading drive submissions:", err.Error())
			return
		}

		fmt.Printf("Found %d submission files. Starting batch evaluation...\n", len(decrypted))

		var records []BatchResultRecord

		for i, sub := range decrypted {
			baseName := filepath.Base(sub.EnvelopePath)
			fmt.Printf("[%d/%d] Processing %s...\n", i+1, len(decrypted), baseName)

			record := BatchResultRecord{
				EnvelopePath: sub.EnvelopePath,
				Status:       "error",
			}

			if sub.Error != nil {
				record.Error = fmt.Sprintf("decryption failed: %v", sub.Error)
				records = append(records, record)
				fmt.Printf("  -> Decryption Error: %v\n", sub.Error)
				continue
			}

			// Write decrypted tar bytes to a temp file
			tempFile, err := os.CreateTemp("", "batch-eval-*.tar")
			if err != nil {
				record.Error = fmt.Sprintf("create temp file: %v", err)
				records = append(records, record)
				fmt.Printf("  -> System Error: %v\n", err)
				continue
			}
			tempPath := tempFile.Name()
			defer os.Remove(tempPath)

			if _, err := tempFile.Write(sub.PlaintextTar); err != nil {
				tempFile.Close()
				record.Error = fmt.Sprintf("write temp file: %v", err)
				records = append(records, record)
				fmt.Printf("  -> System Error: %v\n", err)
				continue
			}
			tempFile.Close()

			// Run evaluation
			resultJSON, err := evaluateSubmissionFile(cmd.Context(), tempPath, evaluationOptions{
				PrivateStore: batchPrivateStore,
				DockerBinary: batchDockerBinary,
				RegistryURL:  batchRegistryURL,
				BearerToken:  batchBearerToken,
			})
			if err != nil {
				record.Error = fmt.Sprintf("evaluation failed: %v", err)
				records = append(records, record)
				fmt.Printf("  -> Evaluation Error: %v\n", err)
				continue
			}

			var evalResult evaluatorcore.EvaluationResult
			if err := json.Unmarshal(resultJSON, &evalResult); err != nil {
				record.Error = fmt.Sprintf("decode result JSON: %v", err)
				records = append(records, record)
				fmt.Printf("  -> Decode Error: %v\n", err)
				continue
			}

			record.StudentID = evalResult.StudentID
			record.LabID = evalResult.LabID
			record.Version = evalResult.Version
			record.Status = evalResult.Status
			record.EarnedPoints = evalResult.EarnedPoints
			record.MaxPoints = evalResult.MaxPoints
			records = append(records, record)

			fmt.Printf("  -> Success: Student=%s Score=%d/%d\n", record.StudentID, record.EarnedPoints, record.MaxPoints)

			// Optionally save individual result JSON to output-dir
			if batchOutputDir != "" {
				if err := os.MkdirAll(batchOutputDir, 0755); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to create output directory %s: %v\n", batchOutputDir, err)
					continue
				}
				fileName := fmt.Sprintf("result-%s-%s.json", record.LabID, record.StudentID)
				if record.StudentID == "" {
					fileName = fmt.Sprintf("result-error-%s.json", strings.TrimSuffix(baseName, filepath.Ext(baseName)))
				}
				outPath := filepath.Join(batchOutputDir, fileName)
				if err := os.WriteFile(outPath, resultJSON, 0644); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to write result file %s: %v\n", outPath, err)
				}
			}
		}

		// Write outputs
		if batchCSVPath != "" {
			if err := writeCSVReport(batchCSVPath, records); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "Error writing CSV report:", err.Error())
			} else {
				fmt.Println("CSV report written to:", batchCSVPath)
			}
		}

		if batchJSONPath != "" {
			if err := writeJSONReport(batchJSONPath, records); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "Error writing JSON report:", err.Error())
			} else {
				fmt.Println("JSON report written to:", batchJSONPath)
			}
		}

		// Print text summary table to stdout
		printSummaryTable(records)
	},
}

func writeCSVReport(path string, records []BatchResultRecord) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"student_id", "lab_id", "version", "earned_points", "max_points", "status", "envelope_file", "error"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, rec := range records {
		row := []string{
			rec.StudentID,
			rec.LabID,
			rec.Version,
			fmt.Sprintf("%d", rec.EarnedPoints),
			fmt.Sprintf("%d", rec.MaxPoints),
			rec.Status,
			filepath.Base(rec.EnvelopePath),
			rec.Error,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func writeJSONReport(path string, records []BatchResultRecord) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func printSummaryTable(records []BatchResultRecord) {
	fmt.Println("\n=================================== BATCH SUMMARY ===================================")
	fmt.Printf("%-15s %-15s %-8s %-12s %-6s\n", "STUDENT ID", "LAB ID", "VERSION", "STATUS", "SCORE")
	fmt.Println("-------------------------------------------------------------------------------------")
	for _, rec := range records {
		scoreStr := fmt.Sprintf("%d/%d", rec.EarnedPoints, rec.MaxPoints)
		if rec.Status == "error" {
			scoreStr = "N/A"
		}
		studentID := rec.StudentID
		if studentID == "" {
			studentID = "[unknown]"
		}
		labID := rec.LabID
		if labID == "" {
			labID = "[unknown]"
		}
		statusStr := rec.Status
		if rec.Error != "" {
			statusStr = "failed"
		}
		fmt.Printf("%-15s %-15s %-8s %-12s %-6s\n", studentID, labID, rec.Version, statusStr, scoreStr)
		if rec.Error != "" {
			fmt.Printf("   Error: %s\n", rec.Error)
		}
	}
	fmt.Println("=====================================================================================")
}

func init() {
	driveCmd.AddCommand(driveEvaluateBatchCmd)

	driveEvaluateBatchCmd.Flags().StringVar(
		&batchRecipientPrivateKey,
		"recipient-private-key",
		"",
		"Base64-encoded X25519 recipient private key used to decrypt the submissions",
	)
	_ = driveEvaluateBatchCmd.MarkFlagRequired("recipient-private-key")

	driveEvaluateBatchCmd.Flags().StringVar(
		&batchPrivateStore,
		"private-store",
		"",
		"Private exercise artifact store to use (defaults to local private cache)",
	)
	driveEvaluateBatchCmd.Flags().StringVar(
		&batchDockerBinary,
		"docker-binary",
		"",
		"Docker binary or Docker host URI used by the evaluator runtime",
	)
	driveEvaluateBatchCmd.Flags().StringVar(
		&batchRegistryURL,
		"registry-url",
		"",
		"Registry server base URL to pull private exercises",
	)
	driveEvaluateBatchCmd.Flags().StringVar(
		&batchBearerToken,
		"bearer-token",
		"",
		"Bearer token used to authenticate with the registry server",
	)
	driveEvaluateBatchCmd.Flags().StringVar(
		&batchCSVPath,
		"csv",
		"",
		"Path to write a summary CSV report of grades",
	)
	driveEvaluateBatchCmd.Flags().StringVar(
		&batchJSONPath,
		"json",
		"",
		"Path to write a summary JSON report of grades",
	)
	driveEvaluateBatchCmd.Flags().StringVar(
		&batchOutputDir,
		"output-dir",
		"",
		"Directory to write individual student evaluation result JSON files",
	)
}
