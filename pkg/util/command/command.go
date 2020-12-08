package command

import (
	"fmt"
	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/pkg/auth"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/spf13/cobra"
	"os"
)

type CommandWrapper struct {
	runFunc (func(cmd *cobra.Command, args []string) error)
}

func WrapRunCommand(runFunc func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		err := runFunc(cmd, args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func RoleAssertion(assertionFunc func() error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		assertionErr := assertionFunc()
		printErrorAndAbortIfNeeded(assertionErr)
	}
}

func NamespacedRoleAssertion(assertionFunc func(namespace string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		kubeClient, err := client.GetClient()
		printErrorAndAbortIfNeeded(auth.GetKubeLoginErrorIfNeeded(err))

		namespaceInfo, err := flags.GetNamespaceToUseFromProjectFlagAndPrintError(cmd, kubeClient)
		printErrorAndAbortIfNeeded(auth.GetKubeLoginErrorIfNeeded(err))

		assertionErr := assertionFunc(namespaceInfo.Namespace)
		printErrorAndAbortIfNeeded(assertionErr)
	}
}

func printErrorAndAbortIfNeeded(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
