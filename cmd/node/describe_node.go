package node

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"io"
	"os"
	"text/tabwriter"

	"github.com/run-ai/runai-cli/pkg/helpers"
	"github.com/run-ai/runai-cli/pkg/nodes"
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
		"CPUs.Utilization",
		"CPUs.Usage",
		"GPUs.Utilization",
		"GPUs.Usage",
		"Mem.Usage",
		"Mem.Utilization",
		"Mem.UsageAndUtilization",
		"GPUMem",
	})

	describeNodeShowGpusFields = ui.EnsureStringPaths(types.GPU{}, []string{
		"IndexID", "Allocated",
	})
)

func DescribeCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:     "node [...NODE_NAME]",
		Aliases: []string{"nodes"},
		Short:   "Display detailed information about nodes in the cluster.",
		ValidArgsFunction: GenNodeNames,
		Example: describeNodeExample,
		PreRun:  commandUtil.RoleAssertion(assertion.AssertViewerRole),
		Run: func(cmd *cobra.Command, args []string) {
			nodeInfos, err := GetNodeInfos(false)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			handleDescribeSpecificNodes(nodeInfos, args...)
		},
	}

	return command
}

func handleDescribeSpecificNodes(nodeInfos *[]nodes.NodeInfo, selectedNodeNames ...string) {
	handleSpecificNodes(nodeInfos, describeNodesPrintFn, selectedNodeNames...)
}

func describeNodesPrintFn(nodeInfos *[]nodes.NodeInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	for i, nodeInfo := range *nodeInfos {
		if i > 0 {
			ui.LineDivider(w)
		}
		describeNodePrintFn(w, &nodeInfo)
	}

	ui.End(w)
	_ = w.Flush()
}

func describeNodePrintFn(w io.Writer, nodeInfo *nodes.NodeInfo) {

	nodeResources := nodeInfo.GetResourcesStatus()
	nodeResourcesConvertor := helpers.NodeResourcesStatusConvertor(nodeResources)

	nodeView := types.NodeView{
		Info:   nodeInfo.GetGeneralInfo(),
		CPUs:   nodeResourcesConvertor.ToCpus(),
		Mem:    nodeResourcesConvertor.ToMemory(),
		GPUs:   nodeResourcesConvertor.ToGpus(),
		GPUMem: nodeResourcesConvertor.ToGpuMemory(),
	}

	err := ui.CreateKeyValuePairs(types.NodeView{}, ui.KeyValuePairsOpt{
		DisplayOpt: ui.DisplayOpt{Hide: append(defaultHiddenFields, describeNodeHiddenFields...)},
	}).Render(w, nodeView).Error()

	if err != nil {
		fmt.Print(err)
	}

	if len(nodeResources.NodeGPUs) > 0 {

		ui.SubTitle(w, "NODE GPUs INFO")

		err = ui.CreateTable(types.GPU{}, ui.TableOpt{
			DisplayOpt: ui.DisplayOpt{
				Show: describeNodeShowGpusFields,
			},
		}).
			Render(w, nodeResources.NodeGPUs).
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

}
