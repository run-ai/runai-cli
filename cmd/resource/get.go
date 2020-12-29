package resource

import (
	"github.com/run-ai/runai-cli/cmd/job"
	"github.com/run-ai/runai-cli/pkg/auth"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	printArgs := job.PrintArgs{}

	var command = &cobra.Command{
		Use:   "get",
		Short: "Display details of a job.",
		PreRun: commandUtil.RoleAssertion(auth.AssertViewerRole),
		Run: func(cmd *cobra.Command, args []string) {
			job.RunDescribeJobDEPRECATED(cmd, printArgs, args[0])
		},
		Args:       cobra.RangeArgs(1, 1),
		Deprecated: "Please use 'runai describe job [job-name]' instead.",
	}

	command.Flags().BoolVarP(&printArgs.ShowEvents, "events", "e", true, "Show events relating to job lifecycle.")
	command.Flags().StringVarP(&printArgs.Output, "output", "o", "", "Output format. One of: json|yaml|wide")
	command.Flags().MarkHidden("events")
	command.Flags().MarkHidden("output")
	return command
}
