package resource

import (
	"fmt"

	"github.com/run-ai/runai-cli/cmd/job"
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	printArgs := job.PrintArgs{}

	deprecationMessage := "'get' command is DEPRECATED, please use 'runai describe job [job-name]' instead."

	var command = &cobra.Command{
		Use:   "get",
		Short: fmt.Sprint("Display details of a job. ", deprecationMessage),
		Run: func(cmd *cobra.Command, args []string) {

			fmt.Print("\n", deprecationMessage, "\n\n")
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				return
			}

			job.RunDescribeJob_DEPRECATED(cmd, printArgs, args[0])
		},
	}

	command.Flags().BoolVarP(&printArgs.ShowEvents, "events", "e", true, "Show events relating to job lifecycle.")
	command.Flags().StringVarP(&printArgs.Output, "output", "o", "", "Output format. One of: json|yaml|wide")
	command.Flags().MarkHidden("events")
	command.Flags().MarkHidden("output")
	return command
}
