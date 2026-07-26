package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	evaluatorcore "TDES/internals/evaluator-core"
	exercisestore "TDES/internals/exercise_store"
	"TDES/internals/remote"

	"github.com/spf13/cobra"
)

var evaluateOutputPath string
var evaluatePrivateStore string
var evaluateDockerBinary string
var evaluateRegistryURL string
var evaluateBearerToken string

var evaluateCmd = &cobra.Command{
	Use:   "evaluate [submission-tar]",
	Short: "Evaluate a submission package using locally cached private artifacts",
	Long:  `Evaluate a submission tar package, print the grading result as JSON, and optionally write the same result to a file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resultJSON, err := evaluateSubmissionFile(cmd.Context(), args[0], evaluationOptions{
			PrivateStore: evaluatePrivateStore,
			DockerBinary: evaluateDockerBinary,
			RegistryURL:  evaluateRegistryURL,
			BearerToken:  evaluateBearerToken,
		})
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "Error evaluating submission:", err.Error())
			return
		}

		if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(resultJSON)); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "Error writing result to terminal:", err.Error())
			return
		}

		if evaluateOutputPath == "" {
			return
		}
		if err := os.WriteFile(evaluateOutputPath, append(resultJSON, '\n'), 0644); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "Error writing result file:", err.Error())
		}
	},
}

type evaluationOptions struct {
	PrivateStore string
	DockerBinary string
	RegistryURL  string
	BearerToken  string
}

func evaluateSubmissionFile(ctx context.Context, submissionPath string, options evaluationOptions) ([]byte, error) {
	registryURL := strings.TrimSpace(options.RegistryURL)
	if registryURL == "" {
		registryURL = strings.TrimSpace(os.Getenv("EUC2_REGISTRY_URL"))
	}
	bearerToken := strings.TrimSpace(options.BearerToken)
	if bearerToken == "" {
		bearerToken = strings.TrimSpace(os.Getenv(remote.BearerTokenEnvVar))
	}

	provider := &privateCacheArtifactProvider{
		storeRoot:   options.PrivateStore,
		registryURL: registryURL,
		bearerToken: bearerToken,
	}
	evaluator, err := evaluatorcore.NewEvaluator(provider, nil)
	if err != nil {
		return nil, err
	}

	result, err := evaluator.EvaluateSubmission(ctx, evaluatorcore.EvaluationRequest{
		SubmissionArchivePath: submissionPath,
		DockerBinary:          options.DockerBinary,
	})
	if err != nil {
		if errors.Is(err, evaluatorcore.ErrExerciseNotFound) {
			return nil, fmt.Errorf("private exercise artifact not found: %w", err)
		}
		return nil, err
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode evaluation result: %w", err)
	}
	return data, nil
}

type privateCacheArtifactProvider struct {
	storeRoot   string
	registryURL string
	bearerToken string
}

func (p *privateCacheArtifactProvider) OpenPrivateArtifact(ctx context.Context, orgID, labID, version string) (io.ReadCloser, error) {
	candidateStores := []string{}
	if p.storeRoot != "" {
		candidateStores = append(candidateStores, p.storeRoot)
	}
	machineCacheDir := exercisestore.GetPrivateCacheDir()
	if machineCacheDir != "" && machineCacheDir != p.storeRoot {
		candidateStores = append(candidateStores, machineCacheDir)
	}

	for _, storeRoot := range candidateStores {
		packagePath, err := exercisestore.ResolveExercisePath(storeRoot, labID, version)
		if err == nil {
			if file, err := os.Open(packagePath); err == nil {
				return file, nil
			}
		}
	}

	// Try fetching from registry if configured
	if p.registryURL != "" {
		fmt.Printf("Private exercise package %s@%s not found in local cache. Attempting to pull from registry...\n", labID, version)

		targetStore := machineCacheDir
		if targetStore == "" {
			targetStore = p.storeRoot
		}

		remoteRef := remote.NewRemote(p.registryURL)
		body, err := remoteRef.FetchPrivateFromRemote(labID, version, orgID, p.bearerToken)
		if err != nil {
			return nil, fmt.Errorf("pull private exercise from registry: %w", err)
		}
		defer body.Close()

		tempFile, err := os.CreateTemp("", "remote-private-exercise-*.tar")
		if err != nil {
			return nil, fmt.Errorf("create temp package: %w", err)
		}
		tempPath := tempFile.Name()
		defer os.Remove(tempPath)

		if _, err := io.Copy(tempFile, body); err != nil {
			tempFile.Close()
			return nil, fmt.Errorf("write remote private package: %w", err)
		}
		if err := tempFile.Close(); err != nil {
			return nil, fmt.Errorf("close temp package: %w", err)
		}

		if err := exercisestore.SavePackage(targetStore, tempPath); err != nil {
			return nil, fmt.Errorf("save remote private package to local cache: %w", err)
		}

		packagePath, err := exercisestore.ResolveExercisePath(targetStore, labID, version)
		if err != nil {
			return nil, err
		}
		file, err := os.Open(packagePath)
		if err != nil {
			return nil, fmt.Errorf("open pulled private package: %w", err)
		}
		return file, nil
	}

	return nil, fmt.Errorf("%w: %s@%s (searched stores: %v)", evaluatorcore.ErrExerciseNotFound, labID, version, candidateStores)
}

func init() {
	rootCmd.AddCommand(evaluateCmd)
	evaluateCmd.Flags().StringVarP(
		&evaluateOutputPath,
		"output",
		"o",
		"",
		"Write the evaluation JSON result to this file",
	)
	evaluateCmd.Flags().StringVar(
		&evaluatePrivateStore,
		"private-store",
		"",
		"Private exercise artifact store to use (defaults to EUC2_PRIVATE_STORE_DIR or the local euc2 private store)",
	)
	evaluateCmd.Flags().StringVar(
		&evaluateDockerBinary,
		"docker-binary",
		"",
		"Docker binary or Docker host URI used by the evaluator runtime",
	)
	evaluateCmd.Flags().StringVar(
		&evaluateRegistryURL,
		"registry-url",
		"",
		"Registry server base URL to pull private exercises (falls back to EUC2_REGISTRY_URL env var)",
	)
	evaluateCmd.Flags().StringVar(
		&evaluateBearerToken,
		"bearer-token",
		"",
		"Bearer token used to authenticate with the registry server (falls back to EUC2_REMOTE_BEARER_TOKEN env var)",
	)
}
