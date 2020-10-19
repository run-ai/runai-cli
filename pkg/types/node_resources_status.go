package types

type NodeResourcesStatus struct {
	Capacity                     ResourceList
	Allocatable                  ResourceList
	Limited                      ResourceList
	Allocated                    ResourceList
	AllocatedGPUsUnits         	 int
	FractionalAllocatedGpuUnits  int
	Requested                    ResourceList
	Usage                        ResourceList
	GpuUnits					 []GPU
}
