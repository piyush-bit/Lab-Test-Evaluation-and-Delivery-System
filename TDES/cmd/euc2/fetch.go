package main

import (
	"TDES/internals/drive"
	initMod "TDES/internals/init"
	"TDES/internals/remote"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type fetchSource struct {
	name      string
	shortHand string
	path      string
}

var fetchSources = []fetchSource{
	{
		name:      "drive",
		shortHand: "d",
	},
	{
		name:      "remote",
		shortHand: "r",
	},
}

var fetchOrgID string

var fetchCmd = &cobra.Command{
	Use:   "fetch [id]",
	Short: "Fetch exercise by ID",
	Long:  `Fetch an exercise by its ID. It's job is to fetch the exercise from the source and save it to the local directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, version := initMod.SplitIdWithVersion(args[0])

		source, err := resolveFetchSource(fetchSources)
		if err != nil {
			cmd.Println("Error resolving fetch source:", err.Error())
			return
		}

		if err := fetchExercise(source, id, version); err != nil {
			cmd.Println("Error fetching from source:", source.name, ":", err.Error())
		}
	},
}

func fetchExercise(source fetchSource, id string, version string) error {
	switch source.name {
	case "drive":
		driveRef := &drive.Drive{Path: source.path}
		return driveRef.FetchFromDrive(id, version)
	case "remote":
		remoteRef := remote.NewRemote(source.path)
		return remoteRef.FetchFromRemote(id, version, fetchOrgID)
	default:
		return fmt.Errorf("unsupported fetch source %q", source.name)
	}
}

func resolveFetchSource(sources []fetchSource) (fetchSource, error) {
	var active []fetchSource

	for _, source := range sources {
		if strings.TrimSpace(source.path) == "" {
			continue
		}
		active = append(active, source)
	}

	if len(active) == 0 {
		return fetchSource{}, fmt.Errorf("exactly one source is required; pass one of --drive or --remote")
	}
	if len(active) > 1 {
		return fetchSource{}, fmt.Errorf("only one fetch source is permitted at a time")
	}

	return active[0], nil
}

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().StringVarP(
		&fetchSources[0].path,
		fetchSources[0].name,
		fetchSources[0].shortHand,
		"",
		"Fetch from the drive source using the provided path",
	)
	fetchCmd.Flags().StringVarP(
		&fetchSources[1].path,
		fetchSources[1].name,
		fetchSources[1].shortHand,
		"",
		"Fetch from the remote source using the provided base URL",
	)
	fetchCmd.Flags().StringVar(
		&fetchOrgID,
		"org-id",
		"",
		"Organization identifier used when fetching from remote",
	)
}
