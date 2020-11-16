// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package node

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/run-ai/runai-cli/pkg/helpers"
	"github.com/run-ai/runai-cli/pkg/nodes"
	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	showDetails bool

	commonTopNodeFields = ui.EnsureStringPaths(types.NodeView{}, []string{
		"Info.Name",
		"GPUs.Capacity",
		"GPUs.Utilization",
		"CPUs.Capacity",
		"CPUs.Utilization",
		"Mem.Capacity",
	})

	tableTopNodeFields = ui.EnsureStringPaths(types.NodeView{}, []string{
		"Mem.UsageAndUtilization",
	})

	detailedTopNodeExtraFields = ui.EnsureStringPaths(types.NodeView{}, []string{
		"CPUs.Usage",
		"Mem.Usage",
		"Mem.Utilization",
		"GPUMem.Capacity",
		"GPUMem.Usage",
		"GPUMem.Utilization",
	})

	topNodeHiddenGpusFields = ui.EnsureStringPaths(types.GPU{}, []string{
		"Allocated",
		"MemoryUsage",
		"MemoryUtilization",
	})
)

func NewTopNodeCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:     "node [...NODE_NAME]",
		Aliases: []string{"nodes"},
		Short:   "Display information about nodes in the cluster.",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {

			nodeInfos, err := getNodeInfos(true)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			handleTopSpecificNodes(nodeInfos, showDetails, args...)
		},
	}

	command.Flags().BoolVarP(&showDetails, "details", "d", false, "Display details")
	return command
}

func handleTopSpecificNodes(nodeInfos *[]nodes.NodeInfo, wide bool, selectedNodeNames ...string) {

	handleSpecificNodes(nodeInfos, func(nodeInfos *[]nodes.NodeInfo) {
		displayTopNodes(nodeInfos, wide, len(selectedNodeNames) == 0)
	}, selectedNodeNames...)

}

func displayTopNodes(nodeInfos *[]nodes.NodeInfo, wide bool, showClusterData bool) {

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	clsData := types.ClusterNodesView{}
	nodeViews := []types.NodeView{}
	nodesToGpus := [][]types.GPU{}

	for _, nodeInfo := range *nodeInfos {

		nodeResources := nodeInfo.GetResourcesStatus()
		nodeResourcesConvertor := helpers.NodeResourcesStatusConvertor(nodeResources)
		nodeView := types.NodeView{
			Info:   nodeInfo.GetGeneralInfo(),
			CPUs:   nodeResourcesConvertor.ToCpus(),
			GPUs:   nodeResourcesConvertor.ToGpus(),
			Mem:    nodeResourcesConvertor.ToMemory(),
			GPUMem: nodeResourcesConvertor.ToGpuMemory(),
		}

		if wide {
			nodesToGpus = append(nodesToGpus, nodeResources.NodeGPUs)
		}

		helpers.AddNodeGPUsToClusterNodes(&clsData, nodeView.Info.Status, nodeView.GPUs)
		nodeViews = append(nodeViews, nodeView)
	}

	if wide {
		displayTopNodeWide(w, nodeViews, nodesToGpus)
	} else {
		showUnhealthyGPUs := clsData.UnhealthyGPUs == 0
		displayTopNodeTable(w, nodeViews, showUnhealthyGPUs)
	}

	_ = w.Flush()
}

func displayTopNodeWide(w io.Writer, nodeViews []types.NodeView, nodesToGPUs [][]types.GPU) {

	showFields := append(commonTopNodeFields, detailedTopNodeExtraFields...)

	for i, nodeView := range nodeViews {
		if i > 0 {
			ui.LineDivider(w)
		}

		err := ui.CreateKeyValuePairs(types.NodeView{}, ui.KeyValuePairsOpt{
			DisplayOpt: ui.DisplayOpt{HideAllByDefault: true, Show: showFields},
		}).Render(w, nodeView).Error()

		if err != nil {
			fmt.Print(err)
		}
		nodeGPUs := nodesToGPUs[i]
		if len(nodeGPUs) > 0 {

			ui.SubTitle(w, "NODE GPUs INFO")

			err = ui.CreateTable(types.GPU{}, ui.TableOpt{
				DisplayOpt: ui.DisplayOpt{
					Hide: topNodeHiddenGpusFields,
				},
			}).
				Render(w, nodeGPUs).
				Error()
			if err != nil {
				fmt.Print(err)
			}
		}
	}
	ui.End(w)
}

func displayTopNodeTable(w io.Writer, rows []types.NodeView, showUnhealthyGpus bool) {
	hiddenFields := defaultHiddenFields
	if !showUnhealthyGpus {
		hiddenFields = append(hiddenFields, unhealthyGpusPath...)
	}

	err := ui.CreateTable(types.NodeView{}, ui.TableOpt{
		DisplayOpt: ui.DisplayOpt{
			HideAllByDefault: true,
			Hide:             hiddenFields,
			Show:             append(commonTopNodeFields, tableTopNodeFields...),
		},
	}).Render(w, rows).Error()

	if err != nil {
		fmt.Print(err)
	}
}
