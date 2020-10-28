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
	TotalGpusMemoryPQ = "totalGpusMemory"
	UsedGpusMemoryPQ  = "usedGpusMemory"
	UsedCpusMemoryPQ  = "usedCpusMemory"
	UsedCpusPQ        = "usedCpus"
	UsedGpusPQ        = "usedGpus"
	GpuIdleTimePQ     = "gpuIdleTime"
	UsedGpuPQ         = "usedGpu"
	GpuUsedByPod	  = "gpuUsedByPod"
	UsedGpuMemoryPQ   = "usedGpuMemory"
	TotalGpuMemoryPQ  = "totalGpuMemory"
)

func NewNodeInfo(node v1.Node, pods []v1.Pod, promNodesMap prom.MetricResultsByQueryName) NodeInfo {
	return NodeInfo{
		Node:           node,
		Pods:           pods,
		PrometheusData: promNodesMap,
	}
}

type NodeInfo struct {
	Node           v1.Node
	Pods           []v1.Pod
	PrometheusData prom.MetricResultsByQueryName
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

	if ni.PrometheusData != nil {
		// set usages
		err := hasError(
			prom.SetFloatFromFirstMetric(&nodeResStatus.Usage.CPUs, ni.PrometheusData, UsedCpusPQ),
			prom.SetFloatFromFirstMetric(&nodeResStatus.Usage.GPUs, ni.PrometheusData, UsedGpusPQ),
			prom.SetFloatFromFirstMetric(&nodeResStatus.Usage.Memory, ni.PrometheusData, UsedCpusMemoryPQ),
			prom.SetFloatFromFirstMetric(&nodeResStatus.Usage.GPUMemory, ni.PrometheusData, UsedGpusMemoryPQ),
			// setFloatPromData(&nodeResStatus.Usage.Storage, p, UsedStoragePQ)

			// set total
			prom.SetFloatFromFirstMetric(&nodeResStatus.Capacity.GPUMemory, ni.PrometheusData, TotalGpusMemoryPQ),
			setGpuUnitsFromPromDataAndPods(&nodeResStatus.GpuUnits, ni.PrometheusData, ni.Pods),
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

func setGpuUnitsFromPromDataAndPods(value *[]types.GPU, data prom.MetricResultsByQueryName, pods []v1.Pod) error {
	result := []types.GPU{}
	metricsValuesByGpus, err := prom.GroupMetrics("gpu", data, GpuIdleTimePQ,UsedGpuPQ, UsedGpuMemoryPQ, TotalGpuMemoryPQ, GpuUsedByPod)

	if err != nil {
		return  err
	}

	fractionAllocatedGpus := util.GetSharedGPUsIndexUsedInPods(pods)

	for gpuIndex, valuesByQueryNames := range metricsValuesByGpus {

		allocated := valuesByQueryNames[GpuUsedByPod]
		fractionAllocated, isFraction := fractionAllocatedGpus[gpuIndex]
		if isFraction {
			allocated = fractionAllocated * 100
		}
		result = append(result, types.GPU {
			IndexID: gpuIndex,
			Allocated: allocated,
			Memory: valuesByQueryNames[TotalGpuMemoryPQ],
			MemoryUsage: valuesByQueryNames[UsedGpuMemoryPQ],
			IdleTime: valuesByQueryNames[GpuIdleTimePQ],
			UTIL: valuesByQueryNames[UsedGpuPQ],
		})
	}

	*value = result
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
