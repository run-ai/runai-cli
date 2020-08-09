package cluster

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/util/command"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func runSetCommand(cmd *cobra.Command, args []string) error {

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

func newSetClusterCommand() *cobra.Command {
	commandWrapper := command.NewCommandWrapper(runSetCommand)
	var command = &cobra.Command{
		Use:   "set [cluster]",
		Short: "Set current cluster",
		Run:   commandWrapper.Run,
		Args:  cobra.RangeArgs(1, 1),
	}

	return command
}
