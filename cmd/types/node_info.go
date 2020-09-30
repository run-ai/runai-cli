package types

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	prom "github.com/run-ai/runai-cli/pkg/prometheus"
	v1 "k8s.io/api/core/v1"
)

// todo
const (
	labelNodeRolePrefix = "node-role.kubernetes.io/"

	// nodeLabelRole specifies the role of a node
	nodeLabelRole = "kubernetes.io/role"

	// prometheus query names
	TotalGpusPQ      = "totalGpus"
	TotalGpuMemoryPQ = "totalGpuMemory"
	TotalCpuMemoryPQ = "totalCpuMemory"
	TotalCpusPQ      = "totalCpus"
	UsedGpuMemoryPQ  = "usedGpuMemory"
	UsedCpuMemoryPQ  = "usedCpuMemory"
	UsedCpusPQ       = "usedCpus"
	UsedGpusPQ       = "usedGpus"
	GPUUtilizationPQ = "gpuUtilization"
	GeneralPQ        = "general"
	ReadyPQ          = "ready"
)

func NewNodeInfo(node v1.Node, pods []v1.Pod, promNodesMap prom.ItemsMap) NodeInfo {
	return NodeInfo{
		Node:           node,
		Pods:           pods,
		PrometheusNode: promNodesMap,
	}
}

type NodeInfo struct {
	Node           v1.Node
	Pods           []v1.Pod
	PrometheusNode prom.ItemsMap
}

func (ni *NodeInfo) GetStatus() NodeStatus {
	if !isNodeReady(ni.Node) {
		return NodeNotReady
	}
	return NodeReady
}

func (ni *NodeInfo) GetGeneralInfo() NodeGeneralInfo {
	return NodeGeneralInfo{
		Name:      ni.Node.Name,
		Role:      strings.Join(findNodeRoles(&ni.Node), ","),
		IPAddress: getNodeInternalAddress(ni.Node),
		Status:    ni.GetStatus(),
	}
}

func (ni *NodeInfo) GetResourcesStatus() NodeResourcesStatus {

	// taken for the old code
	// node := nodeInfo.node
	// totalGPU := totalGpuInNode(node)
	// allocatableGPU := allocatableGpuInNode(node)
	// total: getTotalNodeCapacityProp(ni, "cpu"),

	// for _, pod := range nodeInfo.pods {
	// 	allocatedGPU += gpuInPod(pod)
	// }

	// fractionalGPUsUsedInNode := len(getGPUsIndexUsedInPods(nodeInfo.pods))
	// allocatedGPU += fractionalGPUsUsedInNode
	// totalGPU += fractionalGPUsUsedInNode

	// misssing gpu memory, allocated gpu

	nodeResStatus := NodeResourcesStatus{}
	podResStatus := PodResourcesStatus{}

	for _, pod := range ni.Pods {
		podResStatus.Add(getPodResourceStatus(pod))
	}

	// adding the kube data
	nodeResStatus.Requested = podResStatus.Requested
	nodeResStatus.Limited = podResStatus.Limited
	// nodeResStatus.GpuIndex =
	nodeResStatus.Capacity.AddKubeResourceList(ni.Node.Status.Capacity)
	nodeResStatus.Allocatable.AddKubeResourceList(ni.Node.Status.Allocatable)

	// adding the prometheus data
	p, ok := ni.PrometheusNode[ni.Node.Name]
	if ok {
		// set usages
		setPromData(&nodeResStatus.Usage.CPUs, p, UsedCpusPQ)
		setPromData(&nodeResStatus.Usage.GPUs, p, UsedGpusPQ)
		setPromData(&nodeResStatus.Usage.Memory, p, UsedCpuMemoryPQ)
		setPromData(&nodeResStatus.Usage.GPUMemory, p, UsedGpuMemoryPQ)
		// setPromData(&nodeResStatus.Usage.Storage, p, UsedStoragePQ)

		// set total
		setPromData(&nodeResStatus.Capacity.GPUs, p, TotalGpusPQ)
		setPromData(&nodeResStatus.Capacity.GPUMemory, p, TotalGpuMemoryPQ)
		setPromData(&nodeResStatus.Capacity.Memory, p, TotalCpuMemoryPQ)
		setPromData(&nodeResStatus.Capacity.CPUs, p, TotalCpusPQ)
		// setPromData(&nodeResStatus.Capacity.Storage, p, UsedStoragePQ)
	}

	return nodeResStatus
}

func (nodeInfo *NodeInfo) IsGPUExclusiveNode() bool {
	value, ok := nodeInfo.Node.Status.Allocatable[NVIDIAGPUResourceName]

	if ok {
		ok = (int(value.Value()) > 0)
	}

	return ok
}

// helpers

func setPromData(num *int64, m map[string][]prom.MetricValue, key string) {
	v, found := m[key]
	if !found {
		return
	}
	fmt.Println("key: %s, value %v", key, v)
	*num = v[1].(int64)
}

func isNodeReady(node v1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// todo: create kube utils
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

// todo: create kube utils
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
