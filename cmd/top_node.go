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

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/run-ai/runai-cli/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/run-ai/runai-cli/pkg/ui"
	"github.com/run-ai/runai-cli/cmd/services"
	"github.com/run-ai/runai-cli/cmd/types"


	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	showDetails bool
)

// requested / allocated / usage / utilization (0-100 | 0-[number of units * 100]) / shortcut



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
			allPods, err = acquireAllActivePods(clientset)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			nd := services.NewNodeDescriber(clientset, allPods)
			nodeInfos, err := nd.GetAllNodeInfos()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			displayTopNode(nodeInfos)
		},
	}

	command.Flags().BoolVarP(&showDetails, "details", "d", false, "Display details")
	return command
}




func displayTopNode(nodes []types.NodeInfo) {
	if showDetails {
		displayTopNodeDetails(nodes)
	} else {
		displayTopNodeSummary(nodes)
	}
}

func displayTopNodeSummary(nodeInfos []types.NodeInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	clsData := types.ClusterNodesView{}
	rows := []types.NodeView{}

	for _, nodeInfo := range nodeInfos {

		nrs := nodeInfo.GetResourcesStatus()
		nodeView := types.NodeView {
			Info: nodeInfo.GetGeneralInfo(),
			CPUs: nrs.GetCpus(),
			GPUs: nrs.GetGpus(),
			Mem: nrs.GetMemory(),
			GPUMem: nrs.GetGpuMemory(),
			Storage: nrs.GetStorage(),
		}

		rows = append(rows, nodeView)
		clsData.AddNode(nodeView.Info.Status, nodeView.GPUs)
	}

	ui.CreateTable(types.NodeView{}, ui.TableOpt {}).Render(w, rows)

	clsData.Render(w)	

	_ = w.Flush()
}


func displayTopNodeDetails(nodeInfos []types.NodeInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	clsData := types.ClusterNodesView{}
	fmt.Fprintf(w, "\n")
	for _, nodeInfo := range nodeInfos {
		
		info := nodeInfo.GetGeneralInfo()

		rs := nodeInfo.GetResourcesStatus()
		gpus := rs.GetGpus()

		clsData.AddNode(info.Status, gpus)

		if len(info.Role) == 0 {
			info.Role = "<none>"
		}

		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "NAME:\t%s\n", info.Name)
		fmt.Fprintf(w, "IPADDRESS:\t%s\n", info.IPAddress)
		fmt.Fprintf(w, "ROLE:\t%s\n", info.Role)

		pods := gpuPods(nodeInfo.Pods)
		if len(pods) > 0 {
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "NAMESPACE\tNAME\tGPU REQUESTS\t \n")
			for _, pod := range pods {
				fmt.Fprintf(w, "%s\t%s\t%s\t\n", pod.Namespace,
					pod.Name,
					strconv.FormatInt(gpuInPod(pod), 10))
			}
			fmt.Fprintf(w, "\n")
		}

		var gpuUsageInNode float64 = 0
		if gpus.Capacity > 0 {
			gpuUsageInNode = float64(gpus.Allocated) / float64(gpus.Capacity) * 100
		} else {
			fmt.Fprintf(w, "\n")
		}

		var gpuUnhealthyPercentageInNode float64 = 0
		if  gpus.Capacity > 0  {
			gpuUnhealthyPercentageInNode = float64(gpus.Unhealthy) / float64(gpus.Capacity) * 100
		}

		fmt.Fprintf(w, "Total GPUs In Node %s:\t%s \t\n", info.Name, strconv.FormatInt(gpus.Capacity, 10))
		fmt.Fprintf(w, "Allocated GPUs In Node %s:\t%s (%d%%)\t\n", info.Name, strconv.FormatInt(gpus.Allocated, 10), int64(gpuUsageInNode))
		if gpus.Unhealthy > 0 {
			fmt.Fprintf(w, "Unhealthy GPUs In Node %s:\t%s (%d%%)\t\n", info.Name, strconv.FormatInt(gpus.Unhealthy, 10), int64(gpuUnhealthyPercentageInNode))

		}
		log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(gpus.Capacity, 10),
			strconv.FormatInt(gpus.Allocated, 10))

		fmt.Fprintf(w, "-----------------------------------------------------------------------------------------\n")
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "\n")
	clsData.Render(w)
	_ = w.Flush()
}




func getRequestedNodeCPU(nodeInfo types.NodeInfo) (AllocatableCPU string) {
	var cpuTotal resource.Quantity
	cpuTotal.Set(0)

	for _, pod := range nodeInfo.Pods {
		for _, container := range pod.Spec.Containers {
			quantity, ok := container.Resources.Requests["cpu"]
			if ok {
				cpuTotal.Add(quantity)
			}
		}
	}

	return fmt.Sprintf("%.1f", float64(cpuTotal.MilliValue())/1000)
}

func getTotalNodeMemory(nodeInfo types.NodeInfo) (totalMemory string) {

	valTotal, ok := nodeInfo.Node.Status.Capacity["memory"]
	if ok {
		return fmt.Sprintf("%dM", valTotal.ScaledValue(resource.Mega))
	}

	return ""
}

func getRequestedNodeMemory(nodeInfo types.NodeInfo) (AllocatableMemory string) {

	var memTotal resource.Quantity
	memTotal.Set(0)

	for _, pod := range nodeInfo.Pods {
		for _, container := range pod.Spec.Containers {
			quantity, ok := container.Resources.Requests["memory"]
			if ok {
				memTotal.Add(quantity)
			}

		}
	}

	return fmt.Sprintf("%dM", memTotal.ScaledValue(resource.Mega))
}

// Does the node have unhealthy GPU
func hasUnhealthyGPU(nodeInfo types.NodeInfo) (unhealthy bool) {
	node := nodeInfo.Node
	totalGPU := totalGpuInNode(node)
	allocatableGPU := allocatableGpuInNode(node)

	unhealthy = totalGPU > allocatableGPU

	if unhealthy {
		log.Debugf("node: %s, allocated GPUs %s, total GPUs %s is unhealthy", nodeInfo.Node.Name, strconv.FormatInt(totalGPU, 10),
			strconv.FormatInt(allocatableGPU, 10))
	}

	return unhealthy
}

func isMasterNode(node v1.Node) bool {
	if _, ok := node.Labels[masterLabelRole]; ok {
		return true
	}

	return false
}
