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
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"github.com/run-ai/runai-cli/cmd/trainer"

	"github.com/run-ai/runai-cli/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/run-ai/runai-cli/pkg/ui"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
)

var (
	showDetails bool
)

// requested / allocated / usage / utilization (0-100 | 0-[number of units * 100]) / shortcut


type NodeStatus string

const (
	NodeReady NodeStatus = "ready"
	NodeNotReady NodeStatus = "notReady"
)

type NodeCPUResource struct {
	Total int64					`title:"TOTAL"`
	Allocated int64				`title:"ALLOCATED"`
	Requested int64				`title:"REQUESTED"`
	Usage float32				`title:"USAGE"`
}

type NodeGPUResource struct {
	Total int64						`title:"TOTAL"`
	Unhealthy int64					`title:"UNHEALTHY"`
	Allocated int64					`title:"ALLOCATED UNITS"`
	AllocatedFraction float32  		`title:"ALLOCATED FRACTION"`
	Requested int64					`title:"REQUESTED"`
	Usage float32					`title:"USAGE"`
}

type NodeMemoryResource struct {
	Mem int64							`format:"memory"`
	MemAllocated int64				`title:"ALLOCATED" format:"memory"`
	MemRequested int64				`title:"REQUESTED" format:"memory"`
	MemUsage int64					`title:"USAGE" format:"memory"`
}

type NodeGeneralInfo struct {
	Name string 					`title:"NAME"`
	IPAddress string 				`title:"IP Address"`
	Role string						`title:"ROLE" def:"<none>"`
	Status NodeStatus				`title:"STATUS"`
}

type NodeView struct {
	Info NodeGeneralInfo            `group:"info" title:"-"`
	CPUs NodeCPUResource			`group:"CPU"`
	GPUs NodeGPUResource			`group:"GPU"`
	Mem NodeMemoryResource			`group:"GPUs MEMORY"`
	GPUMem NodeMemoryResource		`group:"GPUs MEMORY"`
}

type ClusterNodesView struct {
	GPUs            	int64
	UnhealthyGPUs   	int64
	AllocatedGPUs   	int64
	GPUsOnReadyNode 	int64
}

type NodeInfo struct {
	node v1.Node
	pods []v1.Pod
}

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
			allPods, err = trainer.AcquireAllActivePods(clientset)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			nd := newNodeDescriber(clientset, allPods)
			nodeInfos, err := nd.getAllNodeInfos()
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

type NodeDescriber struct {
	client  kubernetes.Interface
	allPods []v1.Pod
}

func newNodeDescriber(client kubernetes.Interface, pods []v1.Pod) *NodeDescriber {
	return &NodeDescriber{
		client:  client,
		allPods: pods,
	}
}

func (d *NodeDescriber) getAllNodeInfos() ([]NodeInfo, error) {
	nodeInfoList := []NodeInfo{}

	nodeList, err := d.client.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		return nodeInfoList, err
	}

	for _, node := range nodeList.Items {

		pods := d.getPodsFromNode(node)
		nodeInfo := NodeInfo{
			node: node,
			pods: pods,
		}
		nodeInfoList = append(nodeInfoList, nodeInfo)
	}
	return nodeInfoList, nil
}

func (d *NodeDescriber) getPodsFromNode(node v1.Node) []v1.Pod {
	pods := []v1.Pod{}
	for _, pod := range d.allPods {
		if pod.Spec.NodeName == node.Name {
			pods = append(pods, pod)
		}
	}

	return pods
}

func displayTopNode(nodes []NodeInfo) {
	if showDetails {
		displayTopNodeDetails(nodes)
	} else {
		displayTopNodeSummary(nodes)
	}
}

func displayTopNodeSummary(nodeInfos []NodeInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	clsData := ClusterNodesView{}
	rows := []NodeView{}

	for _, nodeInfo := range nodeInfos {

		nodeView := NodeView {
			Info: nodeInfo.GetGeneralInfo(),
			CPUs: nodeInfo.GetCpus(),
			GPUs: nodeInfo.GetGpus(),
			Mem: nodeInfo.GetMemory(),
			GPUMem: nodeInfo.GetGpuMemory(),
		}

		rows = append(rows, nodeView)
		clsData.AddNode(nodeView.Info.Status, nodeView.GPUs)
	}

	ui.CreateTable(NodeView{}, ui.TableOpt {}).Render(w, rows)

	clsData.Render(w)	

	_ = w.Flush()
}


func displayTopNodeDetails(nodeInfos []NodeInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	clsData := ClusterNodesView{}
	fmt.Fprintf(w, "\n")
	for _, nodeInfo := range nodeInfos {
		
		info := nodeInfo.GetGeneralInfo()

		gpus := nodeInfo.GetGpus()

		clsData.AddNode(info.Status, gpus)

		if len(info.Role) == 0 {
			info.Role = "<none>"
		}

		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "NAME:\t%s\n", info.Name)
		fmt.Fprintf(w, "IPADDRESS:\t%s\n", info.IPAddress)
		fmt.Fprintf(w, "ROLE:\t%s\n", info.Role)

		pods := gpuPods(nodeInfo.pods)
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
		if gpus.Total > 0 {
			gpuUsageInNode = float64(gpus.Allocated) / float64(gpus.Total) * 100
		} else {
			fmt.Fprintf(w, "\n")
		}

		var gpuUnhealthyPercentageInNode float64 = 0
		if  gpus.Total > 0  {
			gpuUnhealthyPercentageInNode = float64(gpus.Unhealthy) / float64(gpus.Total) * 100
		}

		fmt.Fprintf(w, "Total GPUs In Node %s:\t%s \t\n", info.Name, strconv.FormatInt(gpus.Total, 10))
		fmt.Fprintf(w, "Allocated GPUs In Node %s:\t%s (%d%%)\t\n", info.Name, strconv.FormatInt(gpus.Allocated, 10), int64(gpuUsageInNode))
		if gpus.Unhealthy > 0 {
			fmt.Fprintf(w, "Unhealthy GPUs In Node %s:\t%s (%d%%)\t\n", info.Name, strconv.FormatInt(gpus.Unhealthy, 10), int64(gpuUnhealthyPercentageInNode))

		}
		log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(gpus.Total, 10),
			strconv.FormatInt(gpus.Allocated, 10))

		fmt.Fprintf(w, "-----------------------------------------------------------------------------------------\n")
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "\n")
	clsData.Render(w)
	_ = w.Flush()
}


func (cnv *ClusterNodesView) Render(w io.Writer) {

	// Printed at the detailed display

	// fmt.Fprintf(w, "Allocated/Total GPUs In Cluster:\t")
	// log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(totalGPUsInCluster, 10),
	// 	strconv.FormatInt(allocatedGPUsInCluster, 10))

	// var gpuUsage float64 = 0
	// if totalGPUsInCluster > 0 {
	// 	gpuUsage = float64(allocatedGPUsInCluster) / float64(totalGPUsInCluster) * 100
	// }
	// fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
	// 	strconv.FormatInt(allocatedGPUsInCluster, 10),
	// 	strconv.FormatInt(totalGPUsInCluster, 10),
	// 	int64(gpuUsage))
	// // fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ...)
	// if hasUnhealthyGPUNode {
	// 	fmt.Fprintf(w, "Unhealthy/Total GPUs In Cluster:\t")
	// 	var gpuUnhealthyPercentage float64 = 0
	// 	if totalGPUsInCluster > 0 {
	// 		gpuUnhealthyPercentage = float64(totalUnhealthyGPUsInCluster) / float64(totalGPUsInCluster) * 100
	// 	}
	// 	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
	// 		strconv.FormatInt(totalUnhealthyGPUsInCluster, 10),
	// 		strconv.FormatInt(totalGPUsInCluster, 10),
	// 		int64(gpuUnhealthyPercentage))
	// }

	if cnv.UnhealthyGPUs > 0 {
		fmt.Fprintf(w, "---------------------------------------------------------------------------------------------------\n")
	} else {
		fmt.Fprintf(w, "-----------------------------------------------------------------------------------------\n")
	}
	fmt.Fprintf(w, "Allocated/Total GPUs In Cluster:\n")
	log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(cnv.GPUs, 10),
		strconv.FormatInt(cnv.AllocatedGPUs, 10))
	var gpuUsage float64 = 0
	if cnv.GPUs > 0 {
		gpuUsage = float64(cnv.AllocatedGPUs) / float64(cnv.GPUs) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		strconv.FormatInt(cnv.AllocatedGPUs, 10),
		strconv.FormatInt(cnv.GPUs, 10),
		int64(gpuUsage))
	if cnv.GPUs != cnv.GPUsOnReadyNode {
		if cnv.GPUsOnReadyNode > 0 {
			gpuUsage = float64(cnv.AllocatedGPUs) / float64(cnv.GPUsOnReadyNode) * 100
		} else {
			gpuUsage = 0
		}
		fmt.Fprintf(w, "Allocated/Total GPUs(Active) In Cluster:\n")
		fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
			strconv.FormatInt(cnv.AllocatedGPUs, 10),
			strconv.FormatInt(cnv.GPUsOnReadyNode, 10),
			int64(gpuUsage))
	}

	if cnv.UnhealthyGPUs > 0 {
		fmt.Fprintf(w, "Unhealthy/Total GPUs In Cluster:\n")
		var gpuUnhealthyPercentage float64 = 0
		if cnv.GPUs > 0 {
			gpuUnhealthyPercentage = float64(cnv.UnhealthyGPUs) / float64(cnv.GPUs) * 100
		}
		fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
			strconv.FormatInt(cnv.UnhealthyGPUs, 10),
			strconv.FormatInt(cnv.GPUs, 10),
			int64(gpuUnhealthyPercentage))
	}
}

func (cnv *ClusterNodesView) AddNode(status NodeStatus, gpu NodeGPUResource) {
	cnv.GPUs += gpu.Total
	cnv.AllocatedGPUs += gpu.Allocated
	cnv.UnhealthyGPUs += gpu.Unhealthy
	if status == NodeReady{
		cnv.GPUsOnReadyNode += gpu.Total
	}
}


func (ni *NodeInfo) GetStatus() NodeStatus {
	if !isNodeReady(ni.node) {
		return NodeNotReady
	} 
	return NodeReady
}

func (ni *NodeInfo) GetGeneralInfo() NodeGeneralInfo {
	return NodeGeneralInfo {
		Name: ni.node.Name,
		Role: strings.Join(findNodeRoles(&ni.node), ","),
		IPAddress: getNodeInternalAddress(ni.node),
		Status: ni.GetStatus(),
	}
}	

func (ni *NodeInfo) GetCpus() NodeCPUResource{
	return NodeCPUResource {}
}

func (ni *NodeInfo) GetGpus() NodeGPUResource {
	// node := nodeInfo.node
	// totalGPU = totalGpuInNode(node)
	// allocatableGPU = allocatableGpuInNode(node)
	// // allocatedGPU = gpuInPod()

	// for _, pod := range nodeInfo.pods {
	// 	allocatedGPU += gpuInPod(pod)
	// }

	// fractionalGPUsUsedInNode := sharedGPUsUsedInNode(nodeInfo)
	// allocatedGPU += fractionalGPUsUsedInNode
	// totalGPU += fractionalGPUsUsedInNode

	// 	getTotalNodeCPU(nodeInfo),
	// 	getRequestedNodeCPU(nodeInfo),
	
	return NodeGPUResource {}
}

func (ni *NodeInfo) GetMemory() NodeMemoryResource{
	// 	getTotalNodeMemory(nodeInfo),
	// 	getRequestedNodeMemory(nodeInfo))
	return NodeMemoryResource {}
}

func (ni *NodeInfo) GetGpuMemory() NodeMemoryResource{
	return NodeMemoryResource {}
}


func getTotalNodeCPU(nodeInfo NodeInfo) (totalCPU string) {

	valTotal, ok := nodeInfo.node.Status.Capacity["cpu"]
	if ok {
		return valTotal.String()
	}
	return ""
}

func getRequestedNodeCPU(nodeInfo NodeInfo) (AllocatableCPU string) {
	var cpuTotal resource.Quantity
	cpuTotal.Set(0)

	for _, pod := range nodeInfo.pods {
		for _, container := range pod.Spec.Containers {
			quantity, ok := container.Resources.Requests["cpu"]
			if ok {
				cpuTotal.Add(quantity)
			}
		}
	}

	return fmt.Sprintf("%.1f", float64(cpuTotal.MilliValue())/1000)
}

func getTotalNodeMemory(nodeInfo NodeInfo) (totalMemory string) {

	valTotal, ok := nodeInfo.node.Status.Capacity["memory"]
	if ok {
		return fmt.Sprintf("%dM", valTotal.ScaledValue(resource.Mega))
	}

	return ""
}

func getRequestedNodeMemory(nodeInfo NodeInfo) (AllocatableMemory string) {

	var memTotal resource.Quantity
	memTotal.Set(0)

	for _, pod := range nodeInfo.pods {
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
func hasUnhealthyGPU(nodeInfo NodeInfo) (unhealthy bool) {
	node := nodeInfo.node
	totalGPU := totalGpuInNode(node)
	allocatableGPU := allocatableGpuInNode(node)

	unhealthy = totalGPU > allocatableGPU

	if unhealthy {
		log.Debugf("node: %s, allocated GPUs %s, total GPUs %s is unhealthy", nodeInfo.node.Name, strconv.FormatInt(totalGPU, 10),
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

func (nodeInfo *NodeInfo) isGPUExclusiveNode() bool {
	value, ok := nodeInfo.node.Status.Allocatable[NVIDIAGPUResourceName]

	if ok {
		ok = (int(value.Value()) > 0)
	}

	return ok
}

// findNodeRoles returns the roles of a given node.
// The roles are determined by looking for:
// * a node-role.kubernetes.io/<role>="" label
// * a kubernetes.io/role="<role>" label
func findNodeRoles(node *v1.Node) []string {
	roles := sets.NewString()
	for k, v := range node.Labels {
		switch {
		case strings.HasPrefix(k, labelNodeRolePrefix):
			if role := strings.TrimPrefix(k, labelNodeRolePrefix); len(role) > 0 {
				roles.Insert(role)
			}

		case k == nodeLabelRole && v != "":
			roles.Insert(v)
		}
	}
	return roles.List()
}

func isNodeReady(node v1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

func getNodeInternalAddress(node v1.Node) string {
	address := "unknown"
	if len(node.Status.Addresses) > 0 {
		//address = nodeInfo.node.Status.Addresses[0].Address
		for _, addr := range node.Status.Addresses {
			if addr.Type == v1.NodeInternalIP {
				address = addr.Address
			}
		}
	}
	return address
}
