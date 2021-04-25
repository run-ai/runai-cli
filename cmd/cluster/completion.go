package cluster

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func GenClusterNames(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {

	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	config, err := configAccess.GetStartingConfig()

	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	result := make([]string, 0, len(config.Contexts))

	for name, _ := range config.Contexts {
		result = append(result, name)
	}

	return result, cobra.ShellCompDirectiveNoFileComp
}

