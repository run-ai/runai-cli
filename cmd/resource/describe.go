package resource

import (
	"github.com/run-ai/runai-cli/cmd/job"
	"github.com/run-ai/runai-cli/cmd/node"
	"github.com/run-ai/runai-cli/cmd/template"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/spf13/cobra"
)

func NewDescribeCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:    "describe",
		Short:  "Display detailed information about resources.",
		PreRun: commandUtil.RoleAssertion(assertion.AssertViewerRole),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(node.DescribeCommand())
	command.AddCommand(job.DescribeCommand())
	command.AddCommand(template.DescribeCommand())

	return command
}
