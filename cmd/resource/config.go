package resource

import (
	"github.com/run-ai/runai-cli/cmd/cluster"
	"github.com/run-ai/runai-cli/cmd/project"
	"github.com/spf13/cobra"
)

func ConfigCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:   "config",
		Short: "Set a current configuration to be used by default.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(project.ConfigureCommand())
	command.AddCommand(cluster.ConfigureCommand())

	return command
}
