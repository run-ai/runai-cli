package node

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

func GenNodeNames(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {

	kubeClient, err := client.GetClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	nodeList, err := kubeClient.GetClientset().CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	result := make([]string, 0, len(nodeList.Items))

	for _ , nodeItem := range nodeList.Items {
		result = append(result, nodeItem.Name)
	}

	return result, cobra.ShellCompDirectiveNoFileComp
}