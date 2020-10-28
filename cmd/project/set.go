package project

import (
	"fmt"

	"github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
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

	var command = &cobra.Command{
		Use:   "set [PROJECT]",
		Short: "Set a default project",
		Run:   commandUtil.WrapRunCommand(runSetCommand),
		Args:  cobra.RangeArgs(1, 1),
	}

	return command
}
