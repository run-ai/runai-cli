package resource

import (
	"github.com/run-ai/runai-cli/cmd/cluster"
	"github.com/run-ai/runai-cli/cmd/job"
	"github.com/run-ai/runai-cli/cmd/node"
	"github.com/run-ai/runai-cli/cmd/project"
	"github.com/run-ai/runai-cli/cmd/template"
	"github.com/spf13/cobra"
)

const (
	listExample = `
# Get list of the jobs from current project
runai list jobs

# Get list of jobs from all projects
runai list jobs -A

# Get list of the nodes
runai list nodes

# Get list of the projects
runai list projects

# Get list of the clusters
runai list clusters

# Get list of the templates
runai list templates
`
)

func NewListCommand() *cobra.Command {
	var allNamespaces bool

	var command = &cobra.Command{
		Use:     "list",
		Short:   "Display resource list. By default displays the job list.",
		Example: listExample,
		Run: func(cmd *cobra.Command, args []string) {
			job.RunJobList(cmd, args, allNamespaces)
		},
	}

	command.Flags().BoolVarP(&allNamespaces, "all-projects", "A", false, "list jobs from all projects")

	// create subcommands
	command.AddCommand(node.ListCommand())
	command.AddCommand(job.ListCommand())
	command.AddCommand(project.ListCommand())
	command.AddCommand(cluster.ListCommand())
	command.AddCommand(template.ListCommand())

	return command
}
