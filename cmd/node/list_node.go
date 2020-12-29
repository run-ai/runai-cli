package node

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/auth"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"os"
	"text/tabwriter"

	"github.com/run-ai/runai-cli/pkg/helpers"
	"github.com/run-ai/runai-cli/pkg/nodes"
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
	showListNodeFields = ui.EnsureStringPaths(types.NodeView{}, []string{
		"Info",
		"CPUs.Allocatable",
		"CPUs.Allocated",
		"Mem.Allocatable",
		"Mem.Allocated",
		"GPUs.Allocatable",
		"GPUs.Free",
	})
)

func ListCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:     "nodes [...NODE_NAME]",
		Aliases: []string{"node"},
		Short:   "List all nodes.",
		Example: listNodeExample,
		PreRun:  commandUtil.RoleAssertion(auth.AssertViewerRole),
		Run: func(cmd *cobra.Command, args []string) {

			nodeInfos, err := getNodeInfos(false)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			handleListSpecificNodes(nodeInfos, args...)

		},
	}

	return command
}

func handleListSpecificNodes(nodeInfos *[]nodes.NodeInfo, selectedNodeNames ...string) {
	handleSpecificNodes(nodeInfos, listNodes, selectedNodeNames...)
}

func listNodes(nodeInfos *[]nodes.NodeInfo) {
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
