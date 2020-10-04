package types

import (
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	log "github.com/sirupsen/logrus"

	prom "github.com/run-ai/runai-cli/pkg/prometheus"
	"github.com/run-ai/runai-cli/cmd/util"

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
	nodeResStatus.Allocated = podResStatus.Requested
	// allocated gpu is the amount of all pod gpu limits
	nodeResStatus.Allocated.GPUs = podResStatus.Limited.GPUs
	nodeResStatus.Limited = podResStatus.Limited
	
	nodeResStatus.Capacity.AddKubeResourceList(ni.Node.Status.Capacity)
	nodeResStatus.Allocatable.AddKubeResourceList(ni.Node.Status.Allocatable)

	nodeResStatus.AllocatedGPUsIndices = util.GetGPUsIndexUsedInPods(ni.Pods)

	// adding the prometheus data
	p, ok := ni.PrometheusNode[ni.Node.Name]
	if ok {
		// set usages
		err := hasError(
			setFloatPromData(&nodeResStatus.Usage.CPUs, p, UsedCpusPQ),
			
			setFloatPromData(&nodeResStatus.Usage.GPUs, p, UsedGpusPQ),
			setFloatPromData(&nodeResStatus.Usage.Memory, p, UsedCpuMemoryPQ),
			setFloatPromData(&nodeResStatus.Usage.GPUMemory, p, UsedGpuMemoryPQ),
			// setFloatPromData(&nodeResStatus.Usage.Storage, p, UsedStoragePQ)

			// set total
			setFloatPromData(&nodeResStatus.Capacity.GPUs, p, TotalGpusPQ),
			setFloatPromData(&nodeResStatus.Capacity.GPUMemory, p, TotalGpuMemoryPQ),
			setFloatPromData(&nodeResStatus.Capacity.Memory, p, TotalCpuMemoryPQ),
			setFloatPromData(&nodeResStatus.Capacity.CPUs, p, TotalCpusPQ),
		)

		if err != nil {
			log.Debugf("Failed to extract prometheus data, %v",err)
		}
		// setFloatPromData(&nodeResStatus.Capacity.Storage, p, UsedStoragePQ)
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

func setIntPromData(num *int64, m map[string][]prom.MetricValue, key string) error {
	v, found := m[key]
	if !found {
		return nil
	}

	n, err := strconv.Atoi(v[1].(string))
	if err != nil {
		return err
	} 
	*num = int64(n)	
	return nil
}

func setFloatPromData(num *float64, m map[string][]prom.MetricValue, key string) error {
	v, found := m[key]
	if !found {
		return nil
	}
	n, err := strconv.ParseFloat(v[1].(string), 64)
	if err != nil {
		return err
	} 
	*num = n
	return nil
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


func hasError(errors ...error) error{
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
