package top

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/run-ai/runai-cli/pkg/helpers"
	nodeService "github.com/run-ai/runai-cli/pkg/services/node"
	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/ui"
)

func handleDisplayTopNode(nodeInfos []nodeService.NodeInfo, nodeName string) {
	nodeNames := []string{}
	var matchNodeInfo *nodeService.NodeInfo

	for i := range nodeInfos {
		nodeNames = append(nodeNames, nodeInfos[i].Node.Name)
		if nodeInfos[i].Node.Name == nodeName {
			matchNodeInfo = &nodeInfos[i]
			break
		}
	}
	if matchNodeInfo != nil {
		displayTopNode(matchNodeInfo)
	} else {
		fmt.Printf(
			`No match found for node '%s'

Available node names:
%s

`, nodeName, "\t"+strings.Join(nodeNames, "\n\t"))
	}
}

func displayTopNode(nodeInfo *nodeService.NodeInfo) {

	nodeResourcesConvertor := helpers.NodeResourcesStatusConvertor(nodeInfo.GetResourcesStatus())

	nodeView := types.NodeView{
		Info:   nodeInfo.GetGeneralInfo(),
		CPUs:   nodeResourcesConvertor.ToCpus(),
		Mem:    nodeResourcesConvertor.ToMemory(),
		GPUs:   nodeResourcesConvertor.ToGpus(),
		GPUMem: nodeResourcesConvertor.ToGpuMemory(),
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	ui.Title(w, "NODE SUMMERY INFO")

	err := ui.CreateKeyValuePairs(types.NodeView{}, ui.KeyValuePairsOpt{
		DisplayOpt: ui.DisplayOpt{Hide: defaultHiddenFields},
	}).Render(w, nodeView).Error()

	if err != nil {
		fmt.Print(err)
	}
	_ = w.Flush()

	// todo: print node's pods list
	// todo: print node's gpus list
}
