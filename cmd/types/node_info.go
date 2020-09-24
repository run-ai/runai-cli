package types

import (
	"strings"
	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/api/core/v1"

)

// todo
const (
	labelNodeRolePrefix = "node-role.kubernetes.io/"

	// nodeLabelRole specifies the role of a node
	nodeLabelRole = "kubernetes.io/role"
)

func NewNodeInfo(node v1.Node, pods []v1.Pod) NodeInfo {
	return NodeInfo {
		Node: node,
		Pods: pods,
	}
}

type NodeInfo struct {
	Node v1.Node
	Pods []v1.Pod
}


func (ni *NodeInfo) GetStatus() NodeStatus {
	if !isNodeReady(ni.Node) {
		return NodeNotReady
	} 
	return NodeReady
}

func (ni *NodeInfo) GetGeneralInfo() NodeGeneralInfo {
	return NodeGeneralInfo {
		Name: ni.Node.Name,
		Role: strings.Join(findNodeRoles(&ni.Node), ","),
		IPAddress: getNodeInternalAddress(ni.Node),
		Status: ni.GetStatus(),
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

	nodeResStatus.Requested = podResStatus.Requested
	nodeResStatus.Limited = podResStatus.Limited
	// nodeResStatus.GpuIndex = 
	nodeResStatus.Capacity.AddKubeResourceList( ni.Node.Status.Capacity)
	nodeResStatus.Allocatable.AddKubeResourceList(ni.Node.Status.Allocatable) 
	return nodeResStatus
}

// helper

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

func (nodeInfo *NodeInfo) IsGPUExclusiveNode() bool {
	value, ok := nodeInfo.Node.Status.Allocatable[NVIDIAGPUResourceName]

	if ok {
		ok = (int(value.Value()) > 0)
	}

	return ok
}