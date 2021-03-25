package types

type NodeResourcesStatus struct {
	Capacity                       ResourceList
	Allocatable                    ResourceList
	Limited                        ResourceList
	Allocated                      ResourceList
	GPUsInUse                      int
	NumberOfFractionalAllocatedGpu int
	GpuType						   string
	Requested                      ResourceList
	Usage                          ResourceList
	NodeGPUs                       []GPU
}
