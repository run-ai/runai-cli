package types

type NodeStatus string

const (
	NodeReady    NodeStatus = "ready"
	NodeNotReady NodeStatus = "notReady"
)

type NodeCPUResource struct {
	Capacity    int     `title:"CAPACITY" def:"0"`
	Allocatable float64 `title:"ALLOCATABLE"`
	Requested   float64 `title:"REQUESTED"`
	// Limit float64				`title:"Limit"`
	Usage float64 `title:"USAGE" format:"%"`
}

type NodeGPUResource struct {
	Capacity          int     `title:"CAPACITY" def:"0"`
	Allocatable       float64 `title:"ALLOCATABLE"`
	Unhealthy         int     `title:"UNHEALTHY"`
	AllocatedUnits    int     `title:"ALLOCATED UNITS"`
	AllocatedFraction float64 `title:"ALLOCATED FRACTION"`
	Usage             float64 `title:"USAGE" format:"%"`
}

type NodeMemoryResource struct {
	Capacity    float64 `title:"CAPACITY" format:"memory" def:"0"`
	Allocatable float64 `title:"ALLOCATABLE" format:"memory"`
	Requested   float64 `title:"REQUESTED" format:"memory"`
	// Limit float64				`title:"Limit"`
	Usage float64 `title:"USAGE" format:"memory"`
}

type NodeGeneralInfo struct {
	Name      string     `title:"NAME"`
	Status    NodeStatus `title:"STATUS"`
	IPAddress string     `title:"IP Address"`
	Role      string     `title:"ROLE" def:"<none>"`
}

type NodeView struct {
	Info   NodeGeneralInfo     `group:"GENERAL,flatten"`
	CPUs   *NodeCPUResource    `group:"CPU"`
	GPUs   *NodeGPUResource    `group:"GPU" def:"<none>"`
	Mem    *NodeMemoryResource `group:"MEMORY"`
	GPUMem *NodeMemoryResource `group:"GPU MEMORY" def:"<none>"`
}

type ClusterNodesView struct {
	GPUs                  int
	UnhealthyGPUs         int
	AllocatedGpuUnits     int
	AllocatedGpuFractions float64
	GPUsOnReadyNode       int
}
