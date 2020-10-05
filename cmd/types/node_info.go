package types

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	prom "github.com/run-ai/runai-cli/pkg/prometheus"
	"github.com/run-ai/runai-cli/cmd/util"

	v1 "k8s.io/api/core/v1"
)

// todo
const (
	
	// prometheus query names
	TotalGpuMemoryPQ = "totalGpuMemory"
	UsedGpuMemoryPQ  = "usedGpuMemory"
	UsedCpuMemoryPQ  = "usedCpuMemory"
	UsedCpusPQ       = "usedCpus"
	UsedGpusPQ       = "usedGpus"
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
		Role:      strings.Join(util.GetNodeRoles(&ni.Node), ","),
		IPAddress: util.GetNodeInternalAddress(ni.Node),
		Status:    ni.GetStatus(),
	}
}

func (ni *NodeInfo) GetResourcesStatus() NodeResourcesStatus {

	nodeResStatus := NodeResourcesStatus{}
	podResStatus := PodResourcesStatus{}

	for _, pod := range ni.Pods {
		podResStatus.Add(GetPodResourceStatus(pod))
	}

	// adding the kube data
	nodeResStatus.Requested = podResStatus.Requested
	nodeResStatus.Allocated = podResStatus.Requested
	nodeResStatus.Allocated.GPUs = podResStatus.Allocated.GPUs
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
			setFloatPromData(&nodeResStatus.Capacity.GPUMemory, p, TotalGpuMemoryPQ),
		)

		if err != nil {
			log.Debugf("Failed to extract prometheus data, %v",err)
		}
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

func hasError(errors ...error) error{
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
