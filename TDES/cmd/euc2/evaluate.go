package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	evaluatorcore "TDES/internals/evaluator-core"
	exercisestore "TDES/internals/exercise_store"

	"github.com/spf13/cobra"
)

var evaluateOutputPath string
var evaluatePrivateStore string
var evaluateDockerBinary string

var evaluateCmd = &cobra.Command{
	Use:   "evaluate [submission-tar]",
	Short: "Evaluate a submission package using locally cached private artifacts",
	Long:  `Evaluate a submission tar package, print the grading result as JSON, and optionally write the same result to a file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resultJSON, err := evaluateSubmissionFile(cmd.Context(), args[0], evaluationOptions{
			PrivateStore: evaluatePrivateStore,
			DockerBinary: evaluateDockerBinary,
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
}

func evaluateSubmissionFile(ctx context.Context, submissionPath string, options evaluationOptions) ([]byte, error) {
	provider := &privateCacheArtifactProvider{storeRoot: options.PrivateStore}
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
	storeRoot string
}

func (p *privateCacheArtifactProvider) OpenPrivateArtifact(_ context.Context, _, labID, version string) (io.ReadCloser, error) {
	storeRoot := p.storeRoot
	if storeRoot == "" {
		storeRoot = exercisestore.GetPrivateCacheDir()
	}

	packagePath, err := exercisestore.ResolveExercisePath(storeRoot, labID, version)
	if err != nil {
		return nil, fmt.Errorf("%w: %s@%s", evaluatorcore.ErrExerciseNotFound, labID, version)
	}

	file, err := os.Open(packagePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s@%s", evaluatorcore.ErrExerciseNotFound, labID, version)
		}
		return nil, fmt.Errorf("open private exercise package: %w", err)
	}
	return file, nil
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
}
