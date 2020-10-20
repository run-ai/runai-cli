
package resource

import (
	"github.com/spf13/cobra"
	"github.com/run-ai/runai-cli/cmd/node"

)


func NewDescribeCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "describe",
		Short: "Display detailed information about Runai resources.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	// create subcommands
	command.AddCommand(node.NewDescribeNodeCommand())

	return command
}
