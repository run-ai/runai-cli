package types

type NodeResourcesStatus struct {
	Capacity                     ResourceList
	Allocatable                  ResourceList
	Limited                      ResourceList
	Allocated                    ResourceList
	GPUsInUse         	 		 int
	FractionalAllocatedGpuUnits  int
	Requested                    ResourceList
	Usage                        ResourceList
	GpuUnits					 []GPU
}
