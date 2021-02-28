package node

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/nodes"
	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/ui"
)

var (
	defaultHiddenFields = ui.EnsureStringPaths(types.NodeView{}, []string{
		"GPUs.InUse",
	})

	unhealthyGpusPath = ui.EnsureStringPaths(types.NodeView{}, []string{
		"GPUs.Unhealthy",
	})
)

func getNodeInfos(shouldQueryMetrics bool) (*[]nodes.NodeInfo, error) {
	kubeClient, err := client.GetClient()
	if err != nil {
		return nil, err
	}

	nodeInfos, warning, err := nodes.GetAllNodeInfos(kubeClient, shouldQueryMetrics)
	if err != nil {
		return nil, err
	} else if len(warning) > 0 {
		fmt.Println(warning)
	}

	return &nodeInfos, nil
}

func handleSpecificNodes(nodeInfos *[]nodes.NodeInfo, displayFunction func(*[]nodes.NodeInfo), selectedNodeNames ...string) {
	nodeNames := []string{}
	matchsNodeInfos := []nodes.NodeInfo{}

	if len(*nodeInfos) == 0 {
		fmt.Println("No available node found in cluster")
		return
	}

	// show all if no node selected
	if len(selectedNodeNames) == 0 {
		displayFunction(nodeInfos)

		return
	}

	for _, nodeInfo := range *nodeInfos {
		nodeNames = append(nodeNames, nodeInfo.Node.Name)
		if ui.Contains(selectedNodeNames, nodeInfo.Node.Name) {
			matchsNodeInfos = append(matchsNodeInfos, nodeInfo)
		}
	}
	if len(matchsNodeInfos) != len(selectedNodeNames) {
		notFoundNodeNames := []string{}

		for _, nodeName := range selectedNodeNames {
			if !ui.Contains(nodeNames, nodeName) {
				notFoundNodeNames = append(notFoundNodeNames, nodeName)
			}
		}
		fmt.Printf(
			`No match found for node(s) '%s'

Available node names:
%s

`, notFoundNodeNames, "\t"+strings.Join(nodeNames, "\n\t"))
	}

	displayFunction(&matchsNodeInfos)
}
