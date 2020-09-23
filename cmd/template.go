package cmd

import (
	"github.com/spf13/cobra"
)

func NewTemplateCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "template",
		Short: "Get information about templates in the cluster.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
			}
		},
	}

	command.AddCommand(NewTemplateListCommand())
	command.AddCommand(NewTemplateGetCommand())

	return command
}
