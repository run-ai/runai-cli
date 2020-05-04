package flags

import (
	"fmt"

	"github.com/kubeflow/arena/cmd/arena/commands/util"
	"github.com/kubeflow/arena/cmd/arena/types"
	"github.com/kubeflow/arena/pkg/client"
	"github.com/spf13/cobra"

	constants "github.com/kubeflow/arena/cmd/arena/commands/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Return the namespace to use on first argument. on second argument get
func GetNamespaceToUseFromProjectFlag(cmd *cobra.Command, kubeClient *client.Client) (types.NamespaceInfo, error) {
	namespaceInfo, err := getNamespaceInfoToUseFromProjectFlag(cmd, kubeClient)

	if err != nil {
		return namespaceInfo, err
	}

	if namespaceInfo.ProjectName == "" {
		fmt.Println("Please set a default project by running ‘runai project set <project-name>’ or use the flag -p to use a specific project.")
	}

	return namespaceInfo, nil
}

func getNamespaceInfoToUseFromProjectFlag(cmd *cobra.Command, kubeClient *client.Client) (types.NamespaceInfo, error) {
	flagValue := getFlagValue(cmd, ProjectFlag)
	if flagValue != "" {
		namespace, err := util.GetNamespaceFromProjectName(flagValue, kubeClient)
		return types.NamespaceInfo{
			Namespace:   namespace,
			ProjectName: flagValue,
		}, err
	}

	namespace := kubeClient.GetDefaultNamespace()
	projectName, err := getProjectRelatedToNamespace(namespace, kubeClient)

	if err != nil {
		return types.NamespaceInfo{
			Namespace:   "",
			ProjectName: "",
		}, err
	}

	return types.NamespaceInfo{
		Namespace:   namespace,
		ProjectName: projectName,
	}, nil
}

func getProjectRelatedToNamespace(namespaceName string, kubeClient *client.Client) (string, error) {
	namespace, err := kubeClient.GetClientset().CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})

	if err != nil {
		return "", err
	}

	if namespace.Labels == nil {
		return "", nil
	}

	return namespace.Labels[constants.RUNAI_QUEUE_LABEL], nil
}

func GetNamespaceToUseFromProjectFlagIncludingAll(cmd *cobra.Command, kubeClient *client.Client, allFlag bool) (types.NamespaceInfo, error) {
	if allFlag {
		return types.NamespaceInfo{
			Namespace:   metav1.NamespaceAll,
			ProjectName: types.ALL_PROJECTS,
		}, nil
	} else {
		namespaceInfo, err := getNamespaceInfoToUseFromProjectFlag(cmd, kubeClient)

		if err != nil {
			return namespaceInfo, err
		}

		if namespaceInfo.ProjectName == "" {
			fmt.Println("Please set a default project by running ‘runai project set <project-name>’ or use the flag -A to view all projects, or use the flag -p to view a specific project.")
		}

		return namespaceInfo, nil
	}
}

func getFlagValue(cmd *cobra.Command, name string) string {
	return cmd.Flags().Lookup(name).Value.String()
}
