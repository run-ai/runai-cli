package services

import (
	"fmt"

	t "github.com/run-ai/runai-cli/cmd/types"
	prom "github.com/run-ai/runai-cli/pkg/prometheus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	promethesNodeLabelID = "node"
	nodePQs              = prom.MultiQueries{
		t.TotalGpuMemoryPQ: `(sum(runai_node_gpu_total_memory * 1024 * 1024) by (node))`,
		t.UsedGpusPQ:       `((sum(runai_gpus_is_running_with_pod2) by (node))) + (sum(runai_used_shared_gpu_per_node) by (node))`,
		t.UsedGpuMemoryPQ:  `(sum(runai_node_gpu_used_memory * 1024 * 1024) by (node))`,
		t.UsedCpuMemoryPQ:  `runai_node_memory_used_bytes`,
		t.UsedCpusPQ:       `runai_node_cpu_utilization * 100`,
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

func (d *NodeDescriber) GetAllNodeInfos() ([]t.NodeInfo, error, error) {
	var warn error
	nodeInfoList := []t.NodeInfo{}

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
		warn = fmt.Errorf("Missing some data. \nresone: Can't access to the prometheus server, \ncause err: %s", err)
	}

	for _, node := range nodeList.Items {
		pods := d.GetPodsFromNode(node)
		nodeInfo := t.NewNodeInfo(
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
	for _, pod := range d.allPods {
		if pod.Spec.NodeName == node.Name {
			pods = append(pods, pod)
		}
	}

	return pods
}
