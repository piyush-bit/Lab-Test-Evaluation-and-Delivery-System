package cmd

import (
	"fmt"

	runtests "euc2/internals/run"

	"github.com/spf13/cobra"
)

var runLocal bool

var runCmd = &cobra.Command{
	Use:   "run [exercise-dir]",
	Short: "Run public tests for an exercise",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exercisePath := "."
		if len(args) > 0 {
			exercisePath = args[0]
		}

		config := runtests.Config{
			ExercisePath: exercisePath,
			Stdout:       cmd.OutOrStdout(),
			Stderr:       cmd.ErrOrStderr(),
		}

		var err error
		if runLocal {
			err = runtests.RunTestsLocal(config)
		} else {
			err = runtests.RunTestsDocker(config)
		}
		if err != nil {
			fmt.Println("Error running tests:", err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&runLocal, "local", false, "run the local entrypoint on the host instead of Docker")
}
