package node

import (
	"strconv"
	"strings"

	"github.com/run-ai/runai-cli/pkg/helpers"
	"github.com/run-ai/runai-cli/pkg/types"

	log "github.com/sirupsen/logrus"

	"github.com/run-ai/runai-cli/cmd/util"
	prom "github.com/run-ai/runai-cli/pkg/prometheus"

	v1 "k8s.io/api/core/v1"
)

const (

	// prometheus query names
	TotalGpuMemoryPQ = "totalGpuMemory"
	UsedGpuMemoryPQ  = "usedGpuMemory"
	UsedCpuMemoryPQ  = "usedCpuMemory"
	UsedCpusPQ       = "usedCpus"
	UsedGpusPQ       = "usedGpus"
)

func NewNodeInfo(node v1.Node, pods []v1.Pod, promNodesMap prom.MetricResultsByItems) NodeInfo {
	return NodeInfo{
		Node:           node,
		Pods:           pods,
		PrometheusNode: promNodesMap,
	}
}

type NodeInfo struct {
	Node           v1.Node
	Pods           []v1.Pod
	PrometheusNode prom.MetricResultsByItems
}

func (ni *NodeInfo) GetStatus() types.NodeStatus {
	if !util.IsNodeReady(ni.Node) {
		return types.NodeNotReady
	}
	return types.NodeReady
}

func (ni *NodeInfo) GetGeneralInfo() types.NodeGeneralInfo {
	return types.NodeGeneralInfo{
		Name:      ni.Node.Name,
		Role:      strings.Join(util.GetNodeRoles(&ni.Node), ","),
		IPAddress: util.GetNodeInternalAddress(ni.Node),
		Status:    ni.GetStatus(),
	}
}

func (ni *NodeInfo) GetResourcesStatus() types.NodeResourcesStatus {

	nodeResStatus := types.NodeResourcesStatus{}
	podResStatus := types.PodResourcesStatus{}

	for _, pod := range ni.Pods {
		helpers.AddToPodResourcesStatus(&podResStatus, helpers.GetPodResourceStatus(pod))
	}

	// adding the kube data
	nodeResStatus.Requested = podResStatus.Requested
	nodeResStatus.Allocated = podResStatus.Requested
	nodeResStatus.Allocated.GPUs = podResStatus.Allocated.GPUs
	nodeResStatus.Limited = podResStatus.Limited

	helpers.AddKubeResourceListToResourceList(&nodeResStatus.Capacity, ni.Node.Status.Capacity)
	// fix the gpus capacity (when there is a job that using fractional gpu the gpu will not appear in the node > status > capacity so we need to override the capacity.gpus  )
	totalGpus := int(util.AllocatableGpuInNode(ni.Node))
	// check that the totalGpues is set
	if totalGpus > int(nodeResStatus.Capacity.GPUs) {
		nodeResStatus.FractionalAllocatedGpuUnits = len(util.GetSharedGPUsIndexUsedInPods(ni.Pods))
		nodeResStatus.Capacity.GPUs = float64(totalGpus)
		// update the allocatable too
		nodeResStatus.Allocatable.GPUs += float64(nodeResStatus.FractionalAllocatedGpuUnits)
	}

	helpers.AddKubeResourceListToResourceList(&nodeResStatus.Allocatable, ni.Node.Status.Allocatable)
	nodeResStatus.GPUsInUse = nodeResStatus.FractionalAllocatedGpuUnits + int(podResStatus.Limited.GPUs)

	// adding the prometheus data
	promDataByNode, ok := ni.PrometheusNode[ni.Node.Name]
	if ok {
		// set usages
		err := hasError(
			setFloatPromData(&nodeResStatus.Usage.CPUs, promDataByNode, UsedCpusPQ),
			setFloatPromData(&nodeResStatus.Usage.GPUs, promDataByNode, UsedGpusPQ),
			setFloatPromData(&nodeResStatus.Usage.Memory, promDataByNode, UsedCpuMemoryPQ),
			setFloatPromData(&nodeResStatus.Usage.GPUMemory, promDataByNode, UsedGpuMemoryPQ),
			// setFloatPromData(&nodeResStatus.Usage.Storage, p, UsedStoragePQ)

			// set total
			setFloatPromData(&nodeResStatus.Capacity.GPUMemory, promDataByNode, TotalGpuMemoryPQ),
		)

		if err != nil {
			log.Debugf("Failed to extract prometheus data, %v", err)
		}
	}

	return nodeResStatus
}

func (nodeInfo *NodeInfo) IsGPUExclusiveNode() bool {
	value, ok := nodeInfo.Node.Status.Allocatable[util.NVIDIAGPUResourceName]

	if ok {
		ok = (value.Value() > 0)
	}

	return ok
}

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

func hasError(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
