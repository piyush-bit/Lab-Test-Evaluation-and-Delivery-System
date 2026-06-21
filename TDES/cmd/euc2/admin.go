package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"TDES/internals/registry"
	"TDES/internals/remote"

	"github.com/spf13/cobra"
)

var (
	adminRemoteURL   string
	adminBearerToken string
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Administrative utility commands for lab test evaluation registry",
}

var adminOnboardStudentsCmd = &cobra.Command{
	Use:   "onboard-students [roster-csv-path]",
	Short: "Onboard student IDs from a roster CSV file to the registry server",
	Long:  `Uploads a list of authorized student IDs to the registry server to populate the course roster and enable TOFU PIN registration.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		csvPath := args[0]

		registryURL := strings.TrimSpace(adminRemoteURL)
		if registryURL == "" {
			registryURL = strings.TrimSpace(os.Getenv("EUC2_REGISTRY_URL"))
		}
		if registryURL == "" {
			fmt.Println("Error: registry URL is required (use --remote or EUC2_REGISTRY_URL env var)")
			return
		}

		bearerToken := strings.TrimSpace(adminBearerToken)
		if bearerToken == "" {
			bearerToken = strings.TrimSpace(os.Getenv(remote.BearerTokenEnvVar))
		}
		if bearerToken == "" {
			fmt.Println("Error: bearer token is required (use --bearer-token or EUC2_REMOTE_BEARER_TOKEN env var)")
			return
		}

		file, err := os.Open(csvPath)
		if err != nil {
			fmt.Println("Error opening roster CSV file:", err.Error())
			return
		}
		defer file.Close()

		bodyBuffer := &bytes.Buffer{}
		writer := multipart.NewWriter(bodyBuffer)

		part, err := writer.CreateFormFile("roster_csv", filepath.Base(csvPath))
		if err != nil {
			fmt.Println("Error creating form file:", err.Error())
			return
		}

		if _, err := io.Copy(part, file); err != nil {
			fmt.Println("Error copying file content:", err.Error())
			return
		}
		writer.Close()

		// Construct URL
		url := strings.TrimSuffix(registryURL, "/") + "/v1/admin/onboard"
		req, err := http.NewRequest(http.MethodPost, url, bodyBuffer)
		if err != nil {
			fmt.Println("Error creating http request:", err.Error())
			return
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+bearerToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request to server:", err.Error())
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error: server returned status %s: %s\n", resp.Status, string(respBody))
			return
		}

		fmt.Println("Successfully onboarded roster. Server response:", string(respBody))
	},
}

var (
	adminGetGradesOrgID    string
	adminGetGradesLabID    string
	adminGetGradesCSVPath  string
	adminGetGradesJSONPath string
)

var adminGetGradesCmd = &cobra.Command{
	Use:   "get-grades",
	Short: "Retrieve student submission grades from the registry server",
	Long:  `Downloads student grades from the registry server, supports filtering by org_id and lab_id, and exports them to CSV, JSON, or prints them as a terminal table.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		registryURL := strings.TrimSpace(adminRemoteURL)
		if registryURL == "" {
			registryURL = strings.TrimSpace(os.Getenv("EUC2_REGISTRY_URL"))
		}
		if registryURL == "" {
			fmt.Println("Error: registry URL is required (use --remote or EUC2_REGISTRY_URL env var)")
			return
		}

		bearerToken := strings.TrimSpace(adminBearerToken)
		if bearerToken == "" {
			bearerToken = strings.TrimSpace(os.Getenv(remote.BearerTokenEnvVar))
		}
		if bearerToken == "" {
			fmt.Println("Error: bearer token is required (use --bearer-token or EUC2_REMOTE_BEARER_TOKEN env var)")
			return
		}

		baseURL := strings.TrimSuffix(registryURL, "/") + "/v1/submissions"
		req, err := http.NewRequest(http.MethodGet, baseURL, nil)
		if err != nil {
			fmt.Println("Error creating request:", err.Error())
			return
		}

		q := req.URL.Query()
		if adminGetGradesOrgID != "" {
			q.Add("org_id", adminGetGradesOrgID)
		}
		if adminGetGradesLabID != "" {
			q.Add("lab_id", adminGetGradesLabID)
		}

		// If user only requests CSV, request CSV directly from server
		if adminGetGradesCSVPath != "" && adminGetGradesJSONPath == "" {
			q.Add("format", "csv")
		}
		req.URL.RawQuery = q.Encode()

		req.Header.Set("Authorization", "Bearer "+bearerToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error communicating with server:", err.Error())
			return
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err.Error())
			return
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error: server returned status %s: %s\n", resp.Status, string(respBody))
			return
		}

		// Handle CSV export
		if adminGetGradesCSVPath != "" {
			var csvContent []byte
			if q.Get("format") == "csv" {
				csvContent = respBody
			} else {
				var submissions []registry.SubmissionEvaluation
				if err := json.Unmarshal(respBody, &submissions); err != nil {
					fmt.Println("Error parsing JSON response for CSV conversion:", err.Error())
					return
				}
				csvContent, err = convertSubmissionsToCSV(submissions)
				if err != nil {
					fmt.Println("Error converting submissions to CSV:", err.Error())
					return
				}
			}
			if err := os.WriteFile(adminGetGradesCSVPath, csvContent, 0644); err != nil {
				fmt.Println("Error writing CSV file:", err.Error())
				return
			}
			fmt.Println("CSV report written to:", adminGetGradesCSVPath)
		}

		// Handle JSON export
		if adminGetGradesJSONPath != "" {
			var jsonContent []byte
			if q.Get("format") != "csv" {
				jsonContent = respBody
			} else {
				fmt.Println("Warning: Cannot write JSON from CSV server response. Re-running without --csv to fetch JSON.")
				return
			}
			if err := os.WriteFile(adminGetGradesJSONPath, jsonContent, 0644); err != nil {
				fmt.Println("Error writing JSON file:", err.Error())
				return
			}
			fmt.Println("JSON report written to:", adminGetGradesJSONPath)
		}

		// Print Clean Table to stdout
		if adminGetGradesCSVPath == "" && adminGetGradesJSONPath == "" {
			if q.Get("format") == "csv" {
				fmt.Println(string(respBody))
			} else {
				var submissions []registry.SubmissionEvaluation
				if err := json.Unmarshal(respBody, &submissions); err != nil {
					fmt.Println("Error decoding server response:", err.Error())
					return
				}
				printSubmissionsTable(submissions)
			}
		}
	},
}

func convertSubmissionsToCSV(submissions []registry.SubmissionEvaluation) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	header := []string{"id", "org_id", "student_id", "lab_id", "version", "status", "earned_points", "max_points", "created_at"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, s := range submissions {
		row := []string{
			s.ID,
			s.OrgID,
			s.StudentID,
			s.LabID,
			s.Version,
			s.Status,
			fmt.Sprintf("%d", s.EarnedPoints),
			fmt.Sprintf("%d", s.MaxPoints),
			s.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	return buf.Bytes(), nil
}

func printSubmissionsTable(submissions []registry.SubmissionEvaluation) {
	fmt.Println("\n================================= REGISTRY GRADES ==================================")
	fmt.Printf("%-15s %-15s %-8s %-12s %-6s %-20s\n", "STUDENT ID", "LAB ID", "VERSION", "STATUS", "SCORE", "SUBMITTED AT")
	fmt.Println("------------------------------------------------------------------------------------")
	for _, s := range submissions {
		scoreStr := fmt.Sprintf("%d/%d", s.EarnedPoints, s.MaxPoints)
		dateStr := s.CreatedAt.Format("2006-01-02 15:04:05")
		fmt.Printf("%-15s %-15s %-8s %-12s %-6s %-20s\n", s.StudentID, s.LabID, s.Version, s.Status, scoreStr, dateStr)
	}
	fmt.Println("====================================================================================")
}

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminOnboardStudentsCmd)
	adminCmd.AddCommand(adminGetGradesCmd)

	adminOnboardStudentsCmd.Flags().StringVarP(
		&adminRemoteURL,
		"remote",
		"r",
		"",
		"Registry server base URL (falls back to EUC2_REGISTRY_URL)",
	)
	adminOnboardStudentsCmd.Flags().StringVar(
		&adminBearerToken,
		"bearer-token",
		"",
		"Instructor authorization token (falls back to EUC2_REMOTE_BEARER_TOKEN)",
	)

	adminGetGradesCmd.Flags().StringVarP(
		&adminRemoteURL,
		"remote",
		"r",
		"",
		"Registry server base URL (falls back to EUC2_REGISTRY_URL)",
	)
	adminGetGradesCmd.Flags().StringVar(
		&adminBearerToken,
		"bearer-token",
		"",
		"Instructor authorization token (falls back to EUC2_REMOTE_BEARER_TOKEN)",
	)
	adminGetGradesCmd.Flags().StringVar(
		&adminGetGradesOrgID,
		"org-id",
		"",
		"Filter results by organization ID",
	)
	adminGetGradesCmd.Flags().StringVar(
		&adminGetGradesLabID,
		"lab-id",
		"",
		"Filter results by lab ID",
	)
	adminGetGradesCmd.Flags().StringVar(
		&adminGetGradesCSVPath,
		"csv",
		"",
		"Path to write CSV report of grades",
	)
	adminGetGradesCmd.Flags().StringVar(
		&adminGetGradesJSONPath,
		"json",
		"",
		"Path to write JSON report of grades",
	)
}
