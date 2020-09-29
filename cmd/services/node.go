package services

import (
	t "github.com/run-ai/runai-cli/cmd/types"
	prom "github.com/run-ai/runai-cli/cmd/util/prometheus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)


var (
	promethesNodeLabelID = "node"
	nodePQs = prom.MultiQueries {
		t.TotalGpusPQ: `(count(runai_gpus_is_running_with_pod2) by (node))`,
		t.UsedGpusPQ: `((sum(runai_gpus_is_running_with_pod2) by (node))) + (sum(runai_used_shared_gpu_per_node) by (node))`,
		t.UsedGpuMemoryPQ: `(sum(runai_node_gpu_used_memory * 1024 * 1024) by (node))`,
		t.TotalGpuMemoryPQ: `(sum(runai_node_gpu_total_memory * 1024 * 1024) by (node))`,
		t.TotalCpuMemoryPQ: `(sum (kube_node_status_capacity{resource="memory"}) by (node))`,
		t.UsedCpuMemoryPQ: `(sum(max(sum(kube_pod_container_resource_requests_memory_bytes) by (node,pod,namespace) or max(kube_pod_init_container_resource_requests{resource="memory",unit="bytes"}) by (node,pod,namespace)) by (node,pod,namespace) * on (pod,namespace) group_left() (kube_pod_status_phase{phase=~"Pending|Running|Unknown"}==1)) by (node))`,
		t.TotalCpusPQ: `(sum (kube_node_status_capacity{resource="cpu"}) by (node))`,
		t.UsedCpusPQ: `(sum(max(sum(kube_pod_container_resource_requests_cpu_cores) by (node,pod,namespace) or max(kube_pod_init_container_resource_requests{resource="cpu",unit="core"}) by (node,pod,namespace)) by (node,pod,namespace) * on (pod,namespace) group_left() (kube_pod_status_phase{phase=~"Pending|Running|Unknown"}==1)) by (node))`,
		t.GPUUtilizationPQ: `((sum(runai_node_gpu_utilization) by (node)) / on (node) (count(runai_node_gpu_utilization) by (node)))`,
		t.GeneralPQ: `sum(kube_node_status_condition) by (node, namespace)`,
		t.ReadyPQ: `sum(kube_node_status_condition{condition="Ready",status="true"}) by (node)`,
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

	// get promituse data
	promData, err := prom.MultipuleQueriesToItemsMap(nodePQs, promethesNodeLabelID)
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




