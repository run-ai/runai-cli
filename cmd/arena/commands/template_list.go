package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/templates"
	"github.com/spf13/cobra"
)

func NewTemplateListCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "Display information about templates.",
		Run: func(cmd *cobra.Command, args []string) {
			kubeClient, err := client.GetClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			clientset := kubeClient.GetClientset()

			templates := templates.NewTemplates(clientset)
			configs, err := templates.ListTemplates()

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			PrintTemplates(configs)
		},
	}

	return command
}

func PrintTemplates(templates []templates.Template) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	labelField := []string{"NAME", "DESCRIPTION"}

	PrintLine(w, labelField...)

	for _, config := range templates {
		configName := config.Name
		if config.IsDefault {
			configName = fmt.Sprintf("%s (default)", config.Name)
		}
		PrintLine(w, configName, config.Description)
	}

	w.Flush()
}
