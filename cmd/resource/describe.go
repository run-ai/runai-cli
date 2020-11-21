package resource

import (
	"github.com/run-ai/runai-cli/cmd/job"
	"github.com/run-ai/runai-cli/cmd/node"
	"github.com/run-ai/runai-cli/cmd/template"
	"github.com/spf13/cobra"
)

func NewDescribeCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "describe",
		Short: "Display detailed information about resources.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(node.DescribeCommand())
	command.AddCommand(job.DescribeCommand())
	command.AddCommand(template.DescribeCommand())

	return command
}
