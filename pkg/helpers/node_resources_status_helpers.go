package helpers

import (
	"github.com/run-ai/runai-cli/pkg/types"
)

type NodeResourcesStatusConvertor types.NodeResourcesStatus

func (c *NodeResourcesStatusConvertor) ToCpus() *types.NodeCPUResource {
	nrs := (*types.NodeResourcesStatus)(c)
	result := types.NodeCPUResource{
		Capacity:    int(nrs.Capacity.CPUs) / 1000,
		Allocatable: nrs.Allocatable.CPUs,
		Allocated:   nrs.Requested.CPUs / 1000,
		Util:       nrs.Usage.CPUs,
	}
	if result.Capacity == 0 {
		return nil
	}
	return &result
}

func (c *NodeResourcesStatusConvertor) ToGpus() *types.NodeGPUResource {
	nrs := (*types.NodeResourcesStatus)(c)
	result := types.NodeGPUResource{
		Capacity:          int(nrs.Capacity.GPUs),
		Allocatable:       nrs.Allocatable.GPUs,
		Unhealthy:         int(nrs.Capacity.GPUs) - int(nrs.Allocatable.GPUs),
		InUse:   	   nrs.GPUsInUse,
		Allocated:        	   nrs.Allocated.GPUs,
		Util:             nrs.Usage.GPUs,
	}
	if result.Capacity == 0 {
		return nil
	}
	return &result
}

func (c *NodeResourcesStatusConvertor) ToMemory() *types.NodeMemoryResource {
	nrs := (*types.NodeResourcesStatus)(c)
	result := types.NodeMemoryResource{
		Capacity:    nrs.Capacity.Memory,
		Allocatable: nrs.Allocatable.Memory,
		Allocated:   nrs.Requested.Memory,
		Usage:       nrs.Usage.Memory,
	}
	if result.Capacity == 0 {
		return nil
	}
	return &result
}

func (c *NodeResourcesStatusConvertor) ToGpuMemory() *types.NodeMemoryResource {
	nrs := (*types.NodeResourcesStatus)(c)
	result := types.NodeMemoryResource{
		Capacity:    nrs.Capacity.GPUMemory,
		Allocatable: nrs.Allocatable.GPUMemory,
		Usage:       nrs.Usage.GPUMemory,
	}
	if result.Capacity == 0 {
		return nil
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
