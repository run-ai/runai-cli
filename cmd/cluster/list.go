package cluster

import (
	"fmt"
	"os"
	"text/tabwriter"

	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func runListCommand(cmd *cobra.Command, args []string) error {

	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	config, err := configAccess.GetStartingConfig()

	if err != nil {
		return err
	}

	currentContext := config.CurrentContext

	fmt.Printf("Configured clusters on this computer are:\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "CLUSTER\tCURRENT PROJECT\n")

	for name, context := range config.Contexts {
		namespace := context.Namespace
		project := ""
		if len(namespace) > 6 && namespace[0:6] == "runai-" {
			project = namespace[6:len(namespace)]
		}

		if name == currentContext {
			fmt.Fprintf(w, "%s (current)\t%s\n", name, project)
		} else {
			fmt.Fprintf(w, "%s\t%s\n", name, project)
		}
	}
	_ = w.Flush()

	return nil
}

func listCommandDEPRECATED() *cobra.Command {

	var command = &cobra.Command{
		Use:        "list",
		Short:      fmt.Sprint("List all avaliable clusters."),
		Run:        commandUtil.WrapRunCommand(runListCommand),
		Deprecated: "Please use: 'runai list cluster' instead",
	}

	return command
}

func ListCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:     "clusters",
		Aliases: []string{"cluster"},
		Short:   "List all available clusters",
		Run:     commandUtil.WrapRunCommand(runListCommand),
	}

	return command
}
