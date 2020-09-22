package project

import (
	"fmt"

	"github.com/run-ai/runai-cli/cmd/arena/commands/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/spf13/cobra"
)

func runSetCommand(cmd *cobra.Command, args []string) error {

	project := args[0]
	kubeClient, err := client.GetClient()

	if err != nil {
		return err
	}

	namespaceToSet, err := util.GetNamespaceFromProjectName(project, kubeClient)

	if err != nil {
		return err
	}

	err = kubeClient.SetDefaultNamespace(namespaceToSet)
	if err != nil {
		return err
	}

	fmt.Printf("Project %s has been set as default project\n", project)
	return nil

}

func newSetProjectCommand() *cobra.Command {
	commandWrapper := command.NewCommandWrapper(runSetCommand)
	var command = &cobra.Command{
		Use:   "set [PROJECT]",
		Short: "Set a default project",
		Run:   commandWrapper.Run,
		Args:  cobra.RangeArgs(1, 1),
	}

	return command
}
