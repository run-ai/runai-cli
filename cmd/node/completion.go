package node

import (
	"github.com/spf13/cobra"
)

func GenNodeNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

	nodeInfos, err := GetNodeInfos(false)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	result := make([]string, 0, len(*nodeInfos))

	for _ ,nodeInfo := range *nodeInfos {
		result = append(result, nodeInfo.Node.Name)
	}

	return result, cobra.ShellCompDirectiveNoFileComp
}