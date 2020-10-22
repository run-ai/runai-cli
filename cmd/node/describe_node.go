package node

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/run-ai/runai-cli/pkg/helpers"
	nodeService "github.com/run-ai/runai-cli/pkg/services/node"
	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/ui"

	"github.com/spf13/cobra"
)


const (
	describeNodeExample = `
# Describe a node
  runai describe node [NODE_NAME]

# Describe all nodes
  runai describe node`
)

func NewDescribeNodeCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:   "node [...NODE_NAME]",
		Short: "Display detailed information about nodes in the cluster.",
		Example: describeNodeExample,
		Run: func(cmd *cobra.Command, args []string) {

			nodeInfos, err := getNodeInfos()
			
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			handleDescribeSpecificNodes(nodeInfos, args...)
		
		},
	}

	return command
}


func handleDescribeSpecificNodes(nodeInfos *[]nodeService.NodeInfo, selectedNodeNames ...string) {
	handleSpecificNodes(nodeInfos, describeNodes, selectedNodeNames...)	
}

func describeNodes(nodeInfos *[]nodeService.NodeInfo) {
	for _, nodeInfo := range *nodeInfos {
		describeNode(&nodeInfo)
	}
}

func describeNode(nodeInfo *nodeService.NodeInfo) {

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
