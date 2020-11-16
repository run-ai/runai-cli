package helpers

import (
	"fmt"

	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/ui"
)

type NodeResourcesStatusConvertor types.NodeResourcesStatus

func (c *NodeResourcesStatusConvertor) ToCpus() *types.NodeCPUResource {
	nrs := (*types.NodeResourcesStatus)(c)
	capacity := int(nrs.Capacity.CPUs) / 1000
	if capacity == 0 {
		return nil
	}
	result := types.NodeCPUResource{
		Capacity:    capacity,
		Allocatable: nrs.Allocatable.CPUs / 1000,
		Allocated:   nrs.Requested.CPUs / 1000,
		Usage:       nrs.Usage.CPUs / 100 * float64(capacity),
		Utilization: nrs.Usage.CPUs,
	}
	return &result
}

func (c *NodeResourcesStatusConvertor) ToGpus() *types.NodeGPUResource {
	nrs := (*types.NodeResourcesStatus)(c)
	capacity := int(nrs.Capacity.GPUs)
	if capacity == 0 {
		return nil
	}
	result := types.NodeGPUResource{
		Capacity:    capacity,
		Allocatable: nrs.Allocatable.GPUs,
		Unhealthy:   int(nrs.Capacity.GPUs) - int(nrs.Allocatable.GPUs),
		Allocated:   nrs.Allocated.GPUs,
		Usage:       nrs.Usage.GPUs / 100 * float64(capacity),
		Utilization: nrs.Usage.GPUs,
		InUse:       nrs.GPUsInUse,
		Free:        int(nrs.Capacity.GPUs) - nrs.GPUsInUse,
	}

	return &result
}

func (c *NodeResourcesStatusConvertor) ToMemory() *types.NodeMemoryResource {
	nrs := (*types.NodeResourcesStatus)(c)
	usageAndUtilization, util := MemoryUsageAndUtilization(nrs.Usage.Memory, nrs.Capacity.Memory)
	if nrs.Capacity.Memory == 0 {
		return nil
	}
	result := types.NodeMemoryResource{
		Capacity:            nrs.Capacity.Memory,
		Allocatable:         nrs.Allocatable.Memory,
		Allocated:           nrs.Requested.Memory,
		Usage:               nrs.Usage.Memory,
		Utilization:         util,
		UsageAndUtilization: usageAndUtilization,
	}
	return &result
}

func (c *NodeResourcesStatusConvertor) ToGpuMemory() *types.NodeMemoryResource {
	nrs := (*types.NodeResourcesStatus)(c)
	usageAndUtilization, util := MemoryUsageAndUtilization(nrs.Usage.GPUMemory, nrs.Capacity.GPUMemory)
	if nrs.Capacity.GPUMemory == 0 {
		return nil
	}
	result := types.NodeMemoryResource{
		Capacity:            nrs.Capacity.GPUMemory,
		Allocatable:         nrs.Allocatable.GPUMemory,
		Usage:               nrs.Usage.GPUMemory,
		Utilization:         util,
		UsageAndUtilization: usageAndUtilization,
	}

	return &result
}

// todo: currently we are not understand enough the storage in kube
// func (nrs *NodeResourcesStatus) GetStorage() NodeStorageResource {
// 	return NodeStorageResource{
// 		Capacity:    c.Capacity.Storage,
// 		Allocatable: c.Allocatable.Storage,
// 		Allocated:   c.Allocatable.Storage,
// 		Limited:     c.Limited.Storage,
// 		Usage:       c.Usage.Storage,
// 		Requested:   c.Requested.Storage,
// 	}
// }

func MemoryUsageAndUtilization(usage, capacity float64) (string, float64) {
	usageAsBytes := ui.ByteCountIEC(int64(usage))
	utilization := (usage / capacity) * 100
	utilizationFormatet, _ := ui.PrecantageFormat(utilization, nil)
	return fmt.Sprintf(
		"%s (%s)",
		usageAsBytes,
		utilizationFormatet,
	), utilization
}
