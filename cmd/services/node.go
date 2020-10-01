package services

import (
	t "github.com/run-ai/runai-cli/cmd/types"
	prom "github.com/run-ai/runai-cli/pkg/prometheus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	promethesNodeLabelID = "node"
	nodePQs              = prom.MultiQueries{
		t.TotalGpusPQ:      `(count(runai_gpus_is_running_with_pod2) by (node))`,
		t.TotalGpuMemoryPQ: `(sum(runai_node_gpu_total_memory * 1024 * 1024) by (node))`,
		t.TotalCpuMemoryPQ: `(sum (kube_node_status_capacity{resource="memory"}) by (node))`,
		t.TotalCpusPQ:      `(sum (kube_node_status_capacity{resource="cpu"}) by (node))`,
		t.UsedGpusPQ:       `((sum(runai_gpus_is_running_with_pod2) by (node))) + (sum(runai_used_shared_gpu_per_node) by (node))`,
		t.UsedGpuMemoryPQ:  `(sum(runai_node_gpu_used_memory * 1024 * 1024) by (node))`,
		t.UsedCpuMemoryPQ:  `runai_node_memory_used_bytes`,
		t.UsedCpusPQ:       `runai_node_cpu_utilization * 100`,
		t.GPUUtilizationPQ: `((sum(runai_node_gpu_utilization) by (node)) / on (node) (count(runai_node_gpu_utilization) by (node)))`,
		// t.GeneralPQ: `sum(kube_node_status_condition) by (node, namespace)`,
		// t.ReadyPQ: `sum(kube_node_status_condition{condition="Ready",status="true"}) by (node)`,
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

func (d *NodeDescriber) GetAllNodeInfos() ([]t.NodeInfo, error) {
	nodeInfoList := []t.NodeInfo{}

	nodeList, err := d.client.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		return nodeInfoList, err
	}

	// get prometheus node resources data
	promClient, err := prom.BuildPromethuseClient(d.client)
	if err != nil {
		return nil, err
	}
	promData, err := promClient.MultipuleQueriesToItemsMap(nodePQs, promethesNodeLabelID)
	if err != nil {
		return nil, err
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
	return nodeInfoList, nil
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
