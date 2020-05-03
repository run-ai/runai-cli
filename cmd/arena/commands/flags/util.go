package flags

import (
	"fmt"

	"github.com/kubeflow/arena/cmd/arena/commands/util"
	"github.com/kubeflow/arena/pkg/client"
	"github.com/spf13/cobra"

	constants "github.com/kubeflow/arena/cmd/arena/commands/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Return the namespace to use on first argument. on second argument get
func GetNamespaceToUseFromProjectFlag(cmd *cobra.Command, kubeClient *client.Client) (string, error) {
	namespace, related, err := getNamespaceToUseFromProjectFlagAndIfRelatedToProject(cmd, kubeClient)

	if err != nil {
		return "", err
	}

	if !related {
		fmt.Println("Please set a default project by running ‘runai project set <project-name>’ or use the flag -p to use a specific project.")
	}

	return namespace, nil
}

func getNamespaceToUseFromProjectFlagAndIfRelatedToProject(cmd *cobra.Command, kubeClient *client.Client) (string, bool, error) {
	flagValue := getFlagValue(cmd, ProjectFlag)
	if flagValue != "" {
		namespace, err := util.GetNamespaceFromProjectName(flagValue, kubeClient)
		return namespace, true, err
	}

	namespace := kubeClient.GetDefaultNamespace()
	related, err := checkIfNamespaceRelatedToProject(namespace, kubeClient)

	if err != nil {
		return "", false, err
	}

	return namespace, related, nil
}

func checkIfNamespaceRelatedToProject(namespaceName string, kubeClient *client.Client) (bool, error) {
	namespace, err := kubeClient.GetClientset().CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})

	if err != nil {
		return false, err
	}

	if namespace.Labels == nil {
		return false, nil
	}

	if namespace.Labels[constants.RUNAI_QUEUE_LABEL] == "" {
		return false, nil
	}

	return true, nil
}

func GetNamespaceToUseFromProjectFlagIncludingAll(cmd *cobra.Command, kubeClient *client.Client, allFlag bool) (string, error) {
	if allFlag {
		return metav1.NamespaceAll, nil
	} else {
		namespace, related, err := getNamespaceToUseFromProjectFlagAndIfRelatedToProject(cmd, kubeClient)

		if err != nil {
			return "", err
		}

		if !related {
			fmt.Println("Please set a default project by running ‘runai project set <project-name>’ or use the flag -A to view all projects, or use the flag -f to view a specific project.")
		}

		return namespace, nil
	}
}

func getFlagValue(cmd *cobra.Command, name string) string {
	return cmd.Flags().Lookup(name).Value.String()
}
