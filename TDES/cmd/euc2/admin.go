package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminOnboardStudentsCmd)

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
}
