package resource

import (
	"github.com/spf13/cobra"
	"github.com/run-ai/runai-cli/cmd/node"
	"github.com/run-ai/runai-cli/cmd/job"

	// podv1 "k8s.io/api/core/v1"
)

const (
	listExample = `
# Get list of the jobs
runai list job

# Get list of jobs from all projects
runai list job -A

# Get list of the nodes
runai list node
`
)


func NewListCommand() *cobra.Command {
	var allNamespaces bool

	var command = &cobra.Command{
		Use:   "list",
		Short: "Display resource list. By default displays the job list.",
		Example: listExample,
		Run: func(cmd *cobra.Command, args []string) {
			job.RunJobList(cmd, args, false)
		},
	}

	command.Flags().BoolVarP(&allNamespaces, "all-projects", "A", false, "list from all projects")

	// create subcommands
	command.AddCommand(node.NewListNodeCommand())
	command.AddCommand(job.NewListJobCommand())

	return command
}