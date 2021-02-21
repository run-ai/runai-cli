package command

import (
	"fmt"
	"os"

	log "github.com/golang/glog"
	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/pkg/authentication/kubeconfig"
	"github.com/spf13/cobra"
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
		var err error
		namespace := flags.GetNamespaceToUseFromProjectFlagOffline(cmd)
		if namespace == "" {
			namespace, err = kubeconfig.GetCurrentContextDefaultNamespace()
			if err != nil {
				log.Error("Please configure which project to use")
			}
		}

		assertionErr := assertionFunc(namespace)
		printErrorAndAbortIfNeeded(assertionErr)
	}
}

func printErrorAndAbortIfNeeded(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
