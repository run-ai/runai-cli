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
	listNodeExample = `
# Get list of the nodes
runai list node

# Get list of specific nodes
runai list node NODE_NAME_1 NODE_NAME_2
`
)

var (
	showListNodeFields = []string{
		"Info",
		"CPUs.Capacity",
		"CPUs.Allocated",
		"Mem.Capacity",
		"Mem.Allocated",
		"GPUs.Capacity",
		"GPUs.Allocated",
	}
)

func NewListNodeCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:     "node [...NODE_NAME]",
		Short:   "List all nodes.",
		Example: listNodeExample,
		Run: func(cmd *cobra.Command, args []string) {

			nodeInfos, err := getNodeInfos()

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			handleListSpecificNodes(nodeInfos, args...)

		},
	}

	return command
}

func handleListSpecificNodes(nodeInfos *[]nodeService.NodeInfo, selectedNodeNames ...string) {
	handleSpecificNodes(nodeInfos, listNodes, selectedNodeNames...)
}

func listNodes(nodeInfos *[]nodeService.NodeInfo) {
	nodeViews := []types.NodeView{}
	for _, nodeInfo := range *nodeInfos {

		nodeResourcesConvertor := helpers.NodeResourcesStatusConvertor(nodeInfo.GetResourcesStatus())

		nodeView := types.NodeView{
			Info:   nodeInfo.GetGeneralInfo(),
			CPUs:   nodeResourcesConvertor.ToCpus(),
			Mem:    nodeResourcesConvertor.ToMemory(),
			GPUs:   nodeResourcesConvertor.ToGpus(),
			GPUMem: nodeResourcesConvertor.ToGpuMemory(),
		}

		nodeViews = append(nodeViews, nodeView)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	err := ui.CreateTable(types.NodeView{}, ui.TableOpt{
		DisplayOpt: ui.DisplayOpt{Show: showListNodeFields},
	}).Render(w, nodeViews).Error()

	ui.End(w)

	if err != nil {
		fmt.Print(err)
	}
	_ = w.Flush()
}
