package node

import (
	"fmt"
	"os"
	"strings"
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

			if len(*nodeInfos) == 0 {
				fmt.Println("No available node found in cluster")
				return
			}

			if len(args) > 0 {
				handleDescribeSpecificNodes(nodeInfos, args...)
			} else {
				
				describeNodes(nodeInfos)
			}
		},
	}

	return command
}


func handleDescribeSpecificNodes(nodeInfos *[]nodeService.NodeInfo, selectedNodeNames ...string) {
	nodeNames := []string{}
	matchsNodeInfos := []*nodeService.NodeInfo{}

	for i := range *nodeInfos {
		nodeInfo := &(*nodeInfos)[i]
		nodeNames = append(nodeNames, nodeInfo.Node.Name)
		if ui.Contains(selectedNodeNames ,nodeInfo.Node.Name )  {
			matchsNodeInfos = append( matchsNodeInfos,nodeInfo)
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

	for i := range matchsNodeInfos {
		describeNode(matchsNodeInfos[i])
	}
	
}

func describeNodes(nodes *[]nodeService.NodeInfo) {
	for i := range *nodes {
		describeNode(&(*nodes)[i])
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

	ui.Title(w, "NODE SUMMERY INFO")

	err := ui.CreateKeyValuePairs(types.NodeView{}, ui.KeyValuePairsOpt{
		DisplayOpt: ui.DisplayOpt{Hide: defaultHiddenFields},
	}).Render(w, nodeView).Error()

	if err != nil {
		fmt.Print(err)
	}

	ui.Title(w, "NODE GPUs INFO")

	err = ui.CreateTable(types.GPU{}, ui.TableOpt{}).
		Render(w, nodeResources.GpuUnits).
		Error()
	if err != nil {
		fmt.Print(err)
	}

	_ = w.Flush()

	// todo: print node's pods list
}
