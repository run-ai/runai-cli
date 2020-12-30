package cluster

import (
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/spf13/cobra"
)

func NewClusterCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:    "cluster",
		Short:  "Cluster-related commands.",
		PreRun: commandUtil.RoleAssertion(assertion.AssertViewerRole),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
			}
		},
		Deprecated: "Please see usage of `runai list clusters` and `runai config cluster` for more information",
	}

	command.AddCommand(listCommandDEPRECATED())
	command.AddCommand(setCommandDEPRECATED())
	return command
}
