package template

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"os"
	"text/tabwriter"

	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/templates"
	"github.com/run-ai/runai-cli/pkg/ui"
	"github.com/spf13/cobra"
)

func ListCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:     "templates",
		Aliases: []string{"template"},
		Short:   "List all templates.",
		PreRun:  commandUtil.RoleAssertion(assertion.AssertViewerRole),
		Run: func(cmd *cobra.Command, args []string) {
			listAllTemplates()
		},
	}

	return command
}

func PrintTemplates(templates []templates.Template) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	labelField := []string{"NAME", "DESCRIPTION"}

	ui.Line(w, labelField...)

	for _, config := range templates {
		configName := config.Name
		if config.IsAdmin {
			configName = fmt.Sprintf("%s (Admin)", config.Name)
		}
		ui.Line(w, configName, config.Description)
	}

	w.Flush()
}

func ListCommandDEPRECATED() *cobra.Command {
	var command = &cobra.Command{
		Use:    "list",
		Short:  "Display information about templates.",
		PreRun: commandUtil.RoleAssertion(assertion.AssertViewerRole),
		Run: func(cmd *cobra.Command, args []string) {
			listAllTemplates()
		},
		Deprecated: "Please see usage of `runai list templates` for more information",
	}

	return command
}

func listAllTemplates() {
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
}
