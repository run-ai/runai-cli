package project

import (
	"github.com/spf13/cobra"
)

func NewProjectCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "project",
		Short: "Project-related commands.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
			}
		},
		Deprecated: "Please see usage of `runai list projects` and `runai config project` for more information",
	}

	command.AddCommand(listCommandDEPRECATED())
	command.AddCommand(setCommandDEPRECATED())
	return command
}
