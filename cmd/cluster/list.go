package cluster

import (
	"context"
	"fmt"
	"github.com/run-ai/runai-cli/cmd/completion"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"text/tabwriter"

	"github.com/run-ai/runai-cli/cmd/constants"
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

	for name, kubeContext := range config.Contexts {

		project := getActiveProjectOfCluster(config, name, kubeContext)

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
		Use:               "clusters",
		Aliases:           []string{"cluster"},
		Short:             "List all available clusters",
		ValidArgsFunction: completion.NoArgs,
		Run:               commandUtil.WrapRunCommand(runListCommand),
	}

	return command
}

//
//   get the active runai project of a given cluster node. in case of any issue, return empty string
//
func getActiveProjectOfCluster(config *api.Config, nodeName string, kubeContext *api.Context) string {

	//
	//   load the configuration of the desired node from kube.config
	//
	nodeConfig := clientcmd.NewNonInteractiveClientConfig(*config, nodeName, &clientcmd.ConfigOverrides{}, nil)

	//
	//   establish a rest client to the cluster node
	//
	restConfig, err := nodeConfig.ClientConfig()
	if err == nil {
		core, err := kubernetes.NewForConfig(restConfig)
		if err == nil {
			//
			//   get the labels of the current namespace. if this is a runai namespace, it should
			//   contain a runai/queue label which holds the name of the project
			//
			result, err := core.CoreV1().Namespaces().Get(context.TODO(), kubeContext.Namespace, v1.GetOptions{})
			if err == nil && result != nil {
				return result.Labels[constants.RunaiQueueLabel]
			}
		}
	}

	return ""
}
