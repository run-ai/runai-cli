package cluster

import (
	"fmt"
	"strings"
	"github.com/run-ai/runai-cli/cmd/completion"
	"os"
	"text/tabwriter"

	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/run-ai/runai-cli/cmd/constants"

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
		project := ""
		if strings.HasPrefix(context.Namespace, constants.RunaiNsProjectPrefix) {
			lenNsPrefix := len(constants.RunaiNsProjectPrefix)
			project = context.Namespace[lenNsPrefix:len(context.Namespace)]
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
		ValidArgsFunction: completion.NoArgs,
		Run:     commandUtil.WrapRunCommand(runListCommand),
	}

	return command
}
