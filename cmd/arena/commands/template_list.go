package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/clusterConfig"
	"github.com/spf13/cobra"
)

func NewTemplateListCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "Display informationo about templates.",
		Run: func(cmd *cobra.Command, args []string) {
			kubeClient, err := client.GetClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			clientset := kubeClient.GetClientset()

			clusterConfigs := clusterConfig.NewClusterConfigs(clientset)
			configs, err := clusterConfigs.ListClusterConfigs()

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			PrintTemplates(configs)
		},
	}

	return command
}

func PrintTemplates(configs []clusterConfig.ClusterConfig) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	labelField := []string{"NAME", "DESCRIPTION"}

	PrintLine(w, labelField...)

	for _, config := range configs {
		configName := config.Name
		if config.IsDefault {
			configName = fmt.Sprintf("%s (default)", config.Name)
		}
		PrintLine(w, configName, config.Description)
	}

	w.Flush()
}
