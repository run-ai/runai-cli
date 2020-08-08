package cluster

import (
	"github.com/spf13/cobra"
)

func NewClusterCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "cluster",
		Short: "Cluster-related commands.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
			}
		},
	}

	command.AddCommand(newListClustersCommand())
	command.AddCommand(newSetClusterCommand())
	return command
}
