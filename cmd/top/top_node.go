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

package top

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/run-ai/runai-cli/cmd/trainer"
	"github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/helpers"
	nodeService "github.com/run-ai/runai-cli/pkg/services/node"
	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/ui"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	showDetails  bool
	defaultHiddenFields = []string{
		"Mem.Allocatable",
		"CPUs.Allocatable",
		"GPUs.Allocatable",
		"GPUMem.Allocatable",
		"GPUMem.Requested",
	}

	generalNodeInfoFields = []string{
		"Info",
	}

	cpuAndMemoryFields = []string{
		"Info.Name",
		"GPUs",
		"GPUMem",
	}

	gpuAndGpuMemoryFields = []string{
		"Info.Name",
		"CPUs",
		"Mem",
	}
)

func NewTopNodeCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:   "node",
		Short: "Display information about nodes in the cluster.",
		Run: func(cmd *cobra.Command, args []string) {
			kubeClient, err := client.GetClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			clientset := kubeClient.GetClientset()
			allPods, err := trainer.AcquireAllActivePods(clientset)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			nd := nodeService.NewNodeDescriber(clientset, allPods)
			nodeInfos, warning, err := nd.GetAllNodeInfos()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else if len(warning) > 0 {
				fmt.Println(warning)
			}

			displayTopNode(nodeInfos)
		},
	}

	command.Flags().BoolVarP(&showDetails, "details", "d", false, "Display details")
	return command
}

func displayTopNode(nodes []nodeService.Info) {
	if showDetails {
		displayTopNodeDetails(nodes)
	} else {
		displayTopNodeSummary(nodes)
	}
}

func displayTopNodeSummary(nodeInfos []nodeService.Info) {

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	clsData := types.ClusterNodesView{}
	rows := []types.NodeView{}

	for _, nodeInfo := range nodeInfos {

		nodeResourcesConvertor := helpers.NodeResourcesStatusConvertor(nodeInfo.GetResourcesStatus())
		nodeView := types.NodeView{
			Info:   nodeInfo.GetGeneralInfo(),
			CPUs:   nodeResourcesConvertor.ToCpus(),
			GPUs:   nodeResourcesConvertor.ToGpus(),
			Mem:    nodeResourcesConvertor.ToMemory(),
			GPUMem: nodeResourcesConvertor.ToGpuMemory(),
		}

		helpers.AddNodeToClusterNodes(&clsData, nodeView.Info.Status, nodeView.GPUs)
		rows = append(rows, nodeView)
	}

	hiddenFields := defaultHiddenFields
	if clsData.UnhealthyGPUs == 0 {
		hiddenFields = append(hiddenFields, "GPUs.Unhealthy")
	}

	ui.Title(w, "GENERAL NODES INFO")
	err := ui.CreateTable(types.NodeView{}, ui.TableOpt{
		Hide: hiddenFields,
		Show: generalNodeInfoFields,
	}).Render(w, rows).Error()

	if err != nil {
		fmt.Print(err)
	}

	ui.Title(w, "CPU & MEMORY NODES INFO")
	err = ui.CreateTable(types.NodeView{}, ui.TableOpt{
		Hide: hiddenFields,
		Show: gpuAndGpuMemoryFields,
	}).Render(w, rows).Error()

	if err != nil {
		fmt.Print(err)
	}

	ui.Title(w, "GPU & GPU MEMORY NODES INFO")
	err = ui.CreateTable(types.NodeView{}, ui.TableOpt{
		Hide: hiddenFields,
		Show: cpuAndMemoryFields,
	}).Render(w, rows).Error()

	if err != nil {
		fmt.Print(err)
	}

	helpers.RenderClusterNodesView(w, clsData)

	ui.End(w)

	_ = w.Flush()
}

func displayTopNodeDetails(nodeInfos []nodeService.Info) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	clsData := types.ClusterNodesView{}
	fmt.Fprintf(w, "\n")
	for _, nodeInfo := range nodeInfos {

		generalNodeInfo := nodeInfo.GetGeneralInfo()

		nodeResourcesConvertor := helpers.NodeResourcesStatusConvertor(nodeInfo.GetResourcesStatus())
		gpus := nodeResourcesConvertor.ToGpus()

		helpers.AddNodeToClusterNodes(&clsData, generalNodeInfo.Status, gpus)

		if len(generalNodeInfo.Role) == 0 {
			generalNodeInfo.Role = "<none>"
		}

		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "NAME:\t%s\n", generalNodeInfo.Name)
		fmt.Fprintf(w, "IPADDRESS:\t%s\n", generalNodeInfo.IPAddress)
		fmt.Fprintf(w, "ROLE:\t%s\n", generalNodeInfo.Role)

		pods := util.GpuPods(nodeInfo.Pods)
		if len(pods) > 0 {
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "NAMESPACE\tNAME\tGPU REQUESTS\t \n")
			for _, pod := range pods {
				fmt.Fprintf(w, "%s\t%s\t%s\t\n", pod.Namespace,
					pod.Name,
					strconv.FormatInt(util.GpuInPod(pod), 10))
			}
			fmt.Fprintf(w, "\n")
		}

		var gpuUsageInNode float64 = 0
		if gpus.Capacity > 0 {
			gpuUsageInNode = float64(gpus.AllocatedUnits) / float64(gpus.Capacity) * 100
		} else {
			fmt.Fprintf(w, "\n")
		}

		var gpuUnhealthyPercentageInNode float64 = 0
		if gpus.Capacity > 0 {
			gpuUnhealthyPercentageInNode = float64(gpus.Unhealthy) / float64(gpus.Capacity) * 100
		}

		fmt.Fprintf(w, "Total GPUs In Node %s:\t%s \t\n", generalNodeInfo.Name, strconv.FormatInt(int64(gpus.Capacity), 10))
		fmt.Fprintf(w, "Allocated GPUs In Node %s:\t%s (%d%%)\t\n", generalNodeInfo.Name, strconv.FormatInt(int64(gpus.AllocatedUnits), 10), int64(gpuUsageInNode))
		if gpus.Unhealthy > 0 {
			fmt.Fprintf(w, "Unhealthy GPUs In Node %s:\t%s (%d%%)\t\n", generalNodeInfo.Name, strconv.FormatInt(int64(gpus.Unhealthy), 10), int64(gpuUnhealthyPercentageInNode))
		}
		log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(int64(gpus.Capacity), 10),
			strconv.FormatInt(int64(gpus.AllocatedUnits), 10))

		fmt.Fprintf(w, "-----------------------------------------------------------------------------------------\n")
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "\n")
	helpers.RenderClusterNodesView(w, clsData)
	_ = w.Flush()
}
