package resource

import (
	"github.com/spf13/cobra"
	"github.com/run-ai/runai-cli/cmd/job"
)


func NewGetCommand() *cobra.Command {
	// depreacted args - belong to the old command > runai get [job_name]
	printArgs := job.PrintArgs{}


	var command = &cobra.Command{
		Use:   "get",
		Short: "Display details of a job. DEPRECATED! use instead > runai describe job",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				return
			}

			// deprecated - belong to the old command > runai get [job_name]
			job.RunDescribeJob_DEPRECATED(cmd, printArgs, args[0])
		},
	}

	// deprecated - belong to the old command > runai get [job_name]
	command.Flags().BoolVarP(&printArgs.ShowEvents, "events", "e", true, "Show events relating to job lifecycle.")
	command.Flags().StringVarP(&printArgs.Output, "output", "o", "", "Output format. One of: json|yaml|wide")
	command.Flags().MarkHidden("events")
	command.Flags().MarkHidden("output")

	// todo: create subcommands (get job, project ...)
	
	return command
}