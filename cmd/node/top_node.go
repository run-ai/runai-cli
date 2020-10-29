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
	nodeService "github.com/run-ai/runai-cli/pkg/services/node"
	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	showDetails bool

	topNodeFields = ui.EnsureStringPaths(types.NodeView{}, []string{
		"Info.Name",
		"Info.Status",
		"GPUs.Capacity",
		"GPUs.Util",
		"CPUs.Capacity",
		"CPUs.Util",
		"Mem.Capacity",
		"Mem.Usage",
	})
)

func NewTopNodeCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:   "node [...NODE_NAME]",
		Aliases: []string{"nodes"},
		Short: "Display information about nodes in the cluster.",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {

			nodeInfos, err := getNodeInfos()
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

func handleTopSpecificNodes(nodeInfos *[]nodeService.NodeInfo, wide bool, selectedNodeNames ...string) {

	handleSpecificNodes(nodeInfos, func(nodeInfos *[]nodeService.NodeInfo) {
		displayTopNodes(nodeInfos, wide, len(selectedNodeNames) == 0)
	}, selectedNodeNames...)

}

func displayTopNodes(nodeInfos *[]nodeService.NodeInfo, wide bool, showClusterData bool) {

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	clsData := types.ClusterNodesView{}
	rows := []types.NodeView{}
	nodesGpuUnits := [][]types.GPU{}

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
			nodesGpuUnits = append(nodesGpuUnits, nodeResources.GpuUnits)
		}

		helpers.AddNodeGPUsToClusterNodes(&clsData, nodeView.Info.Status, nodeView.GPUs)
		rows = append(rows, nodeView)
	}

	hiddenFields := defaultHiddenFields
	if clsData.UnhealthyGPUs == 0 {
		hiddenFields = append(hiddenFields, "GPUs.Unhealthy")
	}

	if wide {
		displayTopNodeWide(w, rows, nodesGpuUnits, hiddenFields)
	} else {
		displayTopNodeTable(w, rows, hiddenFields)
	}

	if showClusterData {
		helpers.RenderClusterNodesView(w, clsData)
	}

	ui.End(w)

	_ = w.Flush()
}

func displayTopNodeWide(w io.Writer, nodeViews []types.NodeView, nodesGpuUnits [][]types.GPU, hiddenFields []string) {

	for i, nodeView := range nodeViews {
		ui.Title(w, nodeView.Info.Name)
		ui.SubTitle(w, "NODE SUMMERY INFO")

		err := ui.CreateKeyValuePairs(types.NodeView{}, ui.KeyValuePairsOpt{
			DisplayOpt: ui.DisplayOpt{HideAllByDefault: true, Show: topNodeFields},
		}).Render(w, nodeView).Error()

		if err != nil {
			fmt.Print(err)
		}
		nodeGpuUnits := nodesGpuUnits[i]
		if len(nodeGpuUnits) > 0 {

			ui.SubTitle(w, "NODE GPUs INFO")

			err = ui.CreateTable(types.GPU{}, ui.TableOpt{
				DisplayOpt: ui.DisplayOpt{
					Hide: []string{"Allocated"},
				},
			}).
				Render(w, nodeGpuUnits).
				Error()
			if err != nil {
				fmt.Print(err)
			}
		}

		ui.End(w)

	}
}

func displayTopNodeTable(w io.Writer, rows []types.NodeView, hiddenFields []string) {
	err := ui.CreateTable(types.NodeView{}, ui.TableOpt{
		DisplayOpt: ui.DisplayOpt{
			HideAllByDefault: true,
			Hide:             hiddenFields,
			Show:             topNodeFields,
		},
	}).Render(w, rows).Error()

	if err != nil {
		fmt.Print(err)
	}
}

