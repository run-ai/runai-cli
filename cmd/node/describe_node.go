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

var (
	describeNodeHiddenFields = ui.EnsureStringPaths(types.NodeView{}, []string{
		"CPUs.Util",
		"GPUs.Util",
		"Mem.Usage",
		"GPUMem.Usage",
	})
)

func NewDescribeNodeCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:     "node [...NODE_NAME]",
		Aliases: []string{"nodes"},
		Short:   "Display detailed information about nodes in the cluster.",
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

	nodeResources := nodeInfo.GetResourcesStatus()
	nodeResourcesConvertor := helpers.NodeResourcesStatusConvertor(nodeResources)

	nodeView := types.NodeView{
		Info:   nodeInfo.GetGeneralInfo(),
		CPUs:   nodeResourcesConvertor.ToCpus(),
		Mem:    nodeResourcesConvertor.ToMemory(),
		GPUs:   nodeResourcesConvertor.ToGpus(),
		GPUMem: nodeResourcesConvertor.ToGpuMemory(),
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	ui.Title(w, nodeView.Info.Name)

	ui.SubTitle(w, "NODE SUMMERY INFO")

	err := ui.CreateKeyValuePairs(types.NodeView{}, ui.KeyValuePairsOpt{
		DisplayOpt: ui.DisplayOpt{Hide: append(defaultHiddenFields, describeNodeHiddenFields...)},
	}).Render(w, nodeView).Error()

	if err != nil {
		fmt.Print(err)
	}

	if len(nodeResources.GpuUnits) > 0 {

		ui.SubTitle(w, "NODE GPUs INFO")

		err = ui.CreateTable(types.GPU{}, ui.TableOpt{
			DisplayOpt: ui.DisplayOpt{
				Hide: []string{"MemoryUsage", "IdleTime"},
			},
		}).
			Render(w, nodeResources.GpuUnits).
			Error()
		if err != nil {
			fmt.Print(err)
		}
	}

	// todo: print node's pods list
	// this is an old code 
	// pods := util.GpuPods(nodeInfo.Pods)
	// if len(pods) > 0 {
	// 	fmt.Fprintf(w, "\n")
	// 	fmt.Fprintf(w, "NAMESPACE\tNAME\tGPU REQUESTS\t \n")
	// 	for _, pod := range pods {
	// 		fmt.Fprintf(w, "%s\t%s\t%s\t\n", pod.Namespace,
	// 			pod.Name,
	// 			strconv.FormatInt(util.GpuInPod(pod), 10))
	// 	}
	// 	fmt.Fprintf(w, "\n")
	// }

	ui.End(w)

	_ = w.Flush()

}
