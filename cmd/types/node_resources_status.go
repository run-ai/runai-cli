package types


type NodeResourcesStatus struct {
	Capacity ResourceList
	Allocatable ResourceList
	Limited ResourceList
	Allocated ResourceList
	AllocatedGPUsIndexes []string
	Requested ResourceList
	Usage ResourceList
}


func (nrs *NodeResourcesStatus) GetCpus() NodeCPUResource {
	return NodeCPUResource {
		Capacity: nrs.Capacity.CPUs,
		Allocatable: nrs.Allocatable.CPUs,
		Allocated: nrs.Allocatable.CPUs,
		Limited: nrs.Limited.CPUs,
		Usage:  nrs.Usage.CPUs,
		Requested: nrs.Requested.CPUs,	
	}
}

func (nrs *NodeResourcesStatus) GetGpus() NodeGPUResource {

	return NodeGPUResource {
		Capacity: nrs.Capacity.GPUs,
		Allocatable: nrs.Allocatable.GPUs,
		// todo: Unhealthy: nrs. ,
		Allocated: int64(len(nrs.AllocatedGPUsIndexes)),
		AllocatedFraction: nrs.Allocatable.GPUs,
		Limited: nrs.Limited.GPUs,
		Usage:  nrs.Usage.GPUs,
		Requested: nrs.Requested.GPUs,	
	}
}

func (nrs *NodeResourcesStatus) GetMemory() NodeMemoryResource{
	return NodeMemoryResource {
		Capacity: nrs.Capacity.Memory,
		Allocatable: nrs.Allocatable.Memory,
		Allocated: nrs.Allocatable.Memory,
		Limited: nrs.Limited.Memory,
		Usage:  nrs.Usage.Memory,
		Requested: nrs.Requested.Memory,	
	}
}

func (nrs *NodeResourcesStatus) GetGpuMemory() NodeMemoryResource{
	return NodeMemoryResource {
		Capacity: nrs.Capacity.GPUMemory,
		Allocatable: nrs.Allocatable.GPUMemory,
		Allocated: nrs.Allocatable.GPUMemory,
		Limited: nrs.Limited.GPUMemory,
		Usage:  nrs.Usage.GPUMemory,
		Requested: nrs.Requested.GPUMemory,	
	}
}


func (nrs *NodeResourcesStatus) GetStorage() NodeStorageResource{
	return NodeStorageResource {
		Capacity: nrs.Capacity.Storage,
		Allocatable: nrs.Allocatable.Storage,
		Allocated: nrs.Allocatable.Storage,
		Limited: nrs.Limited.Storage,
		Usage:  nrs.Usage.Storage,
		Requested: nrs.Requested.Storage,	
	}
}


