package main

import (
	"TDES/internals/drive"
	"TDES/internals/remote"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
var submitPin string
var submitUpdatePin string

type LocalConfig struct {
	StudentID string `json:"student_id"`
	OrgID     string `json:"org_id"`
	Pin       string `json:"pin"`
}

func loadLocalConfig() (LocalConfig, string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return LocalConfig{}, ""
	}
	configDir := filepath.Join(home, ".euc2")
	configPath := filepath.Join(configDir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return LocalConfig{}, configPath
	}

	var config LocalConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return LocalConfig{}, configPath
	}
	return config, configPath
}

func saveLocalConfig(configPath string, config LocalConfig) {
	if configPath == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(configPath), 0755)
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(configPath, data, 0600)
}

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

		config, configPath := loadLocalConfig()

		orgID := strings.TrimSpace(submitOrgID)
		if orgID == "" {
			orgID = strings.TrimSpace(config.OrgID)
		}
		studentID := strings.TrimSpace(submitStudentID)
		if studentID == "" {
			studentID = strings.TrimSpace(config.StudentID)
		}

		if studentID == "" {
			fmt.Print("Please enter your Student ID: ")
			fmt.Scanln(&studentID)
			studentID = strings.TrimSpace(studentID)
			if studentID == "" {
				cmd.Println("Error: Student ID is required")
				return
			}
		}
		if orgID == "" {
			fmt.Print("Please enter your Organization ID (default 'default'): ")
			fmt.Scanln(&orgID)
			orgID = strings.TrimSpace(orgID)
			if orgID == "" {
				orgID = "default"
			}
		}

		pin := strings.TrimSpace(submitPin)
		if pin == "" {
			pin = strings.TrimSpace(config.Pin)
		}

		// Only prompt for PIN when submitting to remote
		if strategy.name == "remote" && pin == "" {
			fmt.Print("No PIN found on this workstation. Please enter a new PIN (min 4 characters) to secure your student ID: ")
			fmt.Scanln(&pin)
			pin = strings.TrimSpace(pin)
			if len(pin) < 4 {
				cmd.Println("Error: PIN must be at least 4 characters long")
				return
			}
		}

		submissionPath, err := submitExercise(strategy, orgID, studentID, pin, submitUpdatePin)
		if err != nil {
			cmd.Println("Error submitting to source:", strategy.name, ":", err.Error())
			return
		}

		// Save config on successful submission
		if configPath != "" {
			config.StudentID = studentID
			config.OrgID = orgID
			if strategy.name == "remote" {
				if submitUpdatePin != "" {
					config.Pin = submitUpdatePin
				} else {
					config.Pin = pin
				}
			}
			saveLocalConfig(configPath, config)
		}

		cmd.Println("Submission result:", submissionPath)
	},
}

func submitExercise(strategy submitStrategy, orgID string, studentID string, pin string, updatePin string) (string, error) {
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
			Pin:          pin,
			NewPin:       updatePin,
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
	submitCmd.Flags().StringVar(
		&submitPin,
		"pin",
		"",
		"Student PIN (overrides cached config pin)",
	)
	submitCmd.Flags().StringVar(
		&submitUpdatePin,
		"update-pin",
		"",
		"Updates current student PIN to this new value on successful submission",
	)
}
