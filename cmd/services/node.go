package services

import (
	"fmt"

	"github.com/run-ai/runai-cli/cmd/types"
	"github.com/run-ai/runai-cli/cmd/util"
	prom "github.com/run-ai/runai-cli/pkg/prometheus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	promethesNodeLabelID = "node"
	nodePQs              = prom.MultiQueries{
		types.TotalGpuMemoryPQ: `(sum(runai_node_gpu_total_memory * 1024 * 1024) by (node))`,
		types.UsedGpusPQ:       `((sum(runai_gpus_is_running_with_pod2) by (node))) + (sum(runai_used_shared_gpu_per_node) by (node))`,
		types.UsedGpuMemoryPQ:  `(sum(runai_node_gpu_used_memory * 1024 * 1024) by (node))`,
		types.UsedCpuMemoryPQ:  `runai_node_memory_used_bytes`,
		types.UsedCpusPQ:       `runai_node_cpu_utilization * 100`,
	}
)

type NodeDescriber struct {
	client  kubernetes.Interface
	allPods []v1.Pod
}

func NewNodeDescriber(client kubernetes.Interface, pods []v1.Pod) *NodeDescriber {
	return &NodeDescriber{
		client:  client,
		allPods: pods,
	}
}

func (d *NodeDescriber) GetAllNodeInfos() ([]types.NodeInfo, error, string) {
	var warn string
	nodeInfoList := []types.NodeInfo{}

	nodeList, err := d.client.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		return nodeInfoList, err, warn
	}

	var promData prom.ItemsMap

	// get prometheus node resources data
	promClient, err := prom.BuildPromethuseClient(d.client)
	if err == nil {
		promData, err = promClient.MultipuleQueriesToItemsMap(nodePQs, promethesNodeLabelID)
	}
	if err != nil {
		warn = fmt.Sprintf("Missing some data. \nreason: Can't access to the prometheus server, \ncause error: %s", err)
	}

	for _, node := range nodeList.Items {
		pods := d.GetPodsFromNode(node)
		nodeInfo := types.NewNodeInfo(
			node,
			pods,
			promData,
		)
		nodeInfoList = append(nodeInfoList, nodeInfo)
	}
	return nodeInfoList, nil, warn
}

func (d *NodeDescriber) GetPodsFromNode(node v1.Node) []v1.Pod {
	pods := []v1.Pod{}
	if !util.IsNodeReady(node) {
		return pods
	}
	for _, pod := range d.allPods {
		if pod.Spec.NodeName == node.Name && pod.Status.Phase == v1.PodRunning{
			pods = append(pods, pod)
		}
	}

	return pods
}
