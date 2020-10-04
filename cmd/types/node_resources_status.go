package types

type NodeResourcesStatus struct {
	Capacity             ResourceList
	Allocatable          ResourceList
	Limited              ResourceList
	Allocated            ResourceList
	AllocatedGPUsIndices []string
	Requested            ResourceList
	Usage                ResourceList
}

func (nrs *NodeResourcesStatus) GetCpus() NodeCPUResource {
	return NodeCPUResource{
		Capacity:    nrs.Capacity.CPUs,
		Allocatable: nrs.Allocatable.CPUs,
		Allocated:   nrs.Allocated.CPUs,
		Usage:       nrs.Usage.CPUs,
	}
}

func (nrs *NodeResourcesStatus) GetGpus() NodeGPUResource {

	return NodeGPUResource{
		Capacity:    nrs.Capacity.GPUs,
		Allocatable: nrs.Allocatable.GPUs,
		// todo: Unhealthy: nrs. ,
		Allocated:         len(nrs.AllocatedGPUsIndices),
		AllocatedFraction: nrs.Allocatable.GPUs,
		Usage:             nrs.Usage.GPUs,
	}
}

func (nrs *NodeResourcesStatus) GetMemory() NodeMemoryResource {
	return NodeMemoryResource{
		Capacity:    nrs.Capacity.Memory,
		Allocatable: nrs.Allocatable.Memory,
		Allocated:   nrs.Allocated.Memory,
		Usage:       nrs.Usage.Memory,
	}
}

func (nrs *NodeResourcesStatus) GetGpuMemory() NodeMemoryResource {
	return NodeMemoryResource{
		Capacity:    nrs.Capacity.GPUMemory,
		Allocatable: nrs.Allocatable.GPUMemory,
		Allocated:   nrs.Allocated.GPUMemory,
		Usage:       nrs.Usage.GPUMemory,
	}
}


// todo: currently we are not understand enough the storage in kube
// func (nrs *NodeResourcesStatus) GetStorage() NodeStorageResource {
// 	return NodeStorageResource{
// 		Capacity:    nrs.Capacity.Storage,
// 		Allocatable: nrs.Allocatable.Storage,
// 		Allocated:   nrs.Allocatable.Storage,
// 		Limited:     nrs.Limited.Storage,
// 		Usage:       nrs.Usage.Storage,
// 		Requested:   nrs.Requested.Storage,
// 	}
// }
