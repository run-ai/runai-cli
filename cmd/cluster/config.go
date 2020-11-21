package cluster

import (
	"fmt"

	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func runConfigCommand(cmd *cobra.Command, args []string) error {

	clusterName := args[0]

	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		fmt.Printf("%s", err)
		return err
	}

	exists := false
	for name := range config.Contexts {
		if clusterName == name {
			exists = true
		}
	}
	if !exists {
		fmt.Printf("Cluster %s does not exist\n", clusterName)
		return nil
	}

	// set current cluster, then modify and save kubeconfig
	config.CurrentContext = clusterName
	err = clientcmd.ModifyConfig(configAccess, *config, true)
	if err != nil {
		fmt.Printf("%s", err)
		return err
	}

	fmt.Printf("Set current cluster to %s \n", clusterName)
	return nil

}

func ConfigureCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:     "cluster [cluster]",
		Aliases: []string{"clusters"},
		Short:   "Configure a default cluster",
		Run:     commandUtil.WrapRunCommand(runConfigCommand),
		Args:    cobra.RangeArgs(1, 1),
	}

	return command
}

func setCommandDEPRECATED() *cobra.Command {

	var command = &cobra.Command{
		Use:        "set [cluster]",
		Short:      fmt.Sprint("Set current cluster."),
		Args:       cobra.RangeArgs(1, 1),
		Run:        commandUtil.WrapRunCommand(runConfigCommand),
		Deprecated: "use: 'runai config cluster' instead",
	}

	return command
}
