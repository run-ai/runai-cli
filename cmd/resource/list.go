package resource

import (
	"github.com/run-ai/runai-cli/cmd/cluster"
	"github.com/run-ai/runai-cli/cmd/job"
	"github.com/run-ai/runai-cli/cmd/node"
	"github.com/run-ai/runai-cli/cmd/project"
	"github.com/spf13/cobra"
)

const (
	listExample = `
# Get list of the jobs
runai list job

# Get list of jobs from all projects
runai list job -A

# Get list of the nodes
runai list node

# Get list of the projects
runai list project

# Get list of the clusters
runai list cluster
`
)

func NewListCommand() *cobra.Command {
	var allNamespaces bool

	var command = &cobra.Command{
		Use:     "list",
		Short:   "Display resource list. By default displays the job list.",
		Example: listExample,
		Run: func(cmd *cobra.Command, args []string) {
			job.RunJobList(cmd, args, false)
		},
	}

	command.Flags().BoolVarP(&allNamespaces, "all-projects", "A", false, "list jobs from all projects")
	command.Flags().MarkDeprecated("all-projects", "please use 'runai list jobs -A' instead.")

	// create subcommands
	command.AddCommand(node.NewListNodeCommand())
	command.AddCommand(job.NewListJobCommand())
	command.AddCommand(project.NewListProjectCommand())
	command.AddCommand(cluster.NewListClusterCommand())

	return command
}
