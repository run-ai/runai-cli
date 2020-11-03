package types

type NodeStatus string

const (
	NodeReady    NodeStatus = "ready"
	NodeNotReady NodeStatus = "notReady"
)

type NodeCPUResource struct {
	Capacity    int     `title:"CAPACITY"`
	Allocatable float64 `title:"ALLOCATABLE"`
	Requested   float64 `title:"REQUESTED"`
	Usage       float64 `title:"USAGE" format:"%"`
}

type NodeGPUResource struct {
	Capacity          int     `title:"CAPACITY"`
	Allocatable       float64 `title:"ALLOCATABLE"`
	Unhealthy         int     `title:"UNHEALTHY"`
	AllocatedUnits    int     `title:"ALLOCATED UNITS"`
	AllocatedFraction float64 `title:"ALLOCATED FRACTION"`
	Usage             float64 `title:"USAGE" format:"%"`
}

type NodeMemoryResource struct {
	Capacity    float64 `title:"CAPACITY" format:"memory"`
	Allocatable float64 `title:"ALLOCATABLE" format:"memory"`
	Requested   float64 `title:"REQUESTED" format:"memory"`
	Usage       float64 `title:"USAGE" format:"memory"`
}

type NodeGeneralInfo struct {
	Name      string     `title:"NAME"`
	Status    NodeStatus `title:"STATUS"`
	IPAddress string     `title:"IP Address"`
	Role      string     `title:"ROLE" def:"<none>"`
}

type NodeView struct {
	Info   NodeGeneralInfo    `group:"GENERAL,flatten"`
	CPUs   NodeCPUResource    `group:"CPU"`
	GPUs   NodeGPUResource    `group:"GPU"`
	Mem    NodeMemoryResource `group:"MEMORY"`
	GPUMem NodeMemoryResource `group:"GPU MEMORY"`
}

type ClusterNodesView struct {
	GPUs                  int
	UnhealthyGPUs         int
	AllocatedGpuUnits     int
	AllocatedGpuFractions float64
	GPUsOnReadyNode       int
}
