package node

import (
	"fmt"

	"github.com/run-ai/runai-cli/cmd/util"
	prom "github.com/run-ai/runai-cli/pkg/prometheus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	promethesNodeLabelID = "node"
	nodePQs              = prom.QueryNameToQuery{
		TotalGpuMemoryPQ: `(sum(runai_node_gpu_total_memory * 1024 * 1024) by (node))`,
		UsedGpusPQ:       `((sum(runai_gpus_is_running_with_pod2) by (node))) + (sum(runai_used_shared_gpu_per_node) by (node))`,
		UsedGpuMemoryPQ:  `(sum(runai_node_gpu_used_memory * 1024 * 1024) by (node))`,
		UsedCpuMemoryPQ:  `runai_node_memory_used_bytes`,
		UsedCpusPQ:       `runai_node_cpu_utilization * 100`,
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

func (d *NodeDescriber) GetAllNodeInfos() ([]NodeInfo, string, error) {
	var warning string
	nodeInfoList := []NodeInfo{}

	nodeList, err := d.client.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		return nodeInfoList, warning, err
	}

	var promData prom.MetricResultsAsItems

	promClient, err := prom.BuildPrometheusClient(d.client)
	if err == nil {
		promData, err = promClient.GroupMultiQueriesToItems(nodePQs, promethesNodeLabelID)
	}
	if err != nil {
		warning = fmt.Sprintf("Missing some data. \nreason: Can't access to the prometheus server, \ncause error: %s", err)
	}

	for _, node := range nodeList.Items {
		pods := d.GetPodsFromNode(node)
		nodeInfo := NewNodeInfo(
			node,
			pods,
			promData,
		)
		nodeInfoList = append(nodeInfoList, nodeInfo)
	}
	return nodeInfoList, warning, err
}

func (d *NodeDescriber) GetPodsFromNode(node v1.Node) []v1.Pod {
	pods := []v1.Pod{}
	if !util.IsNodeReady(node) {
		return pods
	}
	for _, pod := range d.allPods {
		if pod.Spec.NodeName == node.Name &&
			(pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodPending) {
			pods = append(pods, pod)
		}
	}

	return pods
}
