package project

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/auth"

	"github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/spf13/cobra"
)

func runConfigProjectCommand(cmd *cobra.Command, args []string) error {

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

func ConfigureCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:     "project [PROJECT]",
		Aliases: []string{"projects"},
		Short:   "Configure a default project.",
		PreRun:  commandUtil.RoleAssertion(auth.AssertViewerRole),
		Run:     commandUtil.WrapRunCommand(runConfigProjectCommand),
		Args:    cobra.RangeArgs(1, 1),
	}

	return command
}

func setCommandDEPRECATED() *cobra.Command {

	var command = &cobra.Command{
		Use:        "set [PROJECT]",
		Short:      fmt.Sprint("Set a default project all available projects."),
		PreRun: 	commandUtil.RoleAssertion(auth.AssertViewerRole),
		Run:        commandUtil.WrapRunCommand(runConfigProjectCommand),
		Args:       cobra.RangeArgs(1, 1),
		Deprecated: "Please use: 'runai config project' instead",
	}

	return command
}
