package node

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai-cli/pkg/ui"
	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/cmd/trainer"
	"github.com/run-ai/runai-cli/pkg/client"
	nodeService "github.com/run-ai/runai-cli/pkg/services/node"

)

var (
	defaultHiddenFields = ui.EnsureStringPaths(types.NodeView{}, []string{
		"Mem.Allocatable",
		"CPUs.Allocatable",
		"GPUs.Allocatable",
		"GPUs.InUse",
		"GPUMem.Allocatable",
		"GPUMem.Requested",
	})
)

func getNodeInfos() (*[]nodeService.NodeInfo, error) {
		kubeClient, err := client.GetClient()
		if err != nil {
			return nil, err
		}
		clientset := kubeClient.GetClientset()
		allPods, err := trainer.AcquireAllActivePods(clientset)
		if err != nil {
			return nil, err
		}
		nd := nodeService.NewNodeDescriber(clientset, allPods)
		nodeInfos, warning, err := nd.GetAllNodeInfos()
		if err != nil {
			return nil, err
		} else if len(warning) > 0 {
			fmt.Println(warning)
		}

		return &nodeInfos, nil
}


func handleSpecificNodes(nodeInfos *[]nodeService.NodeInfo, displayFunction func(*[]nodeService.NodeInfo)  ,selectedNodeNames ...string) {
	nodeNames := []string{}
	matchsNodeInfos := []nodeService.NodeInfo{}


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
		if ui.Contains(selectedNodeNames ,nodeInfo.Node.Name )  {
			matchsNodeInfos = append( matchsNodeInfos, nodeInfo)
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