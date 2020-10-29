package types



type NodeStatus string

const (
	NodeReady    NodeStatus = "ready"
	NodeNotReady NodeStatus = "notReady"
)

type NodeCPUResource struct {
	Capacity    int     `title:"CAPACITY" def:"0"`
	Allocatable float64 `title:"ALLOCATABLE"`
	Allocated   float64 `title:"ALLOCATED"`
	Util        float64 `title:"Util." format:"%"`
}

type NodeGPUResource struct {
	Capacity    int     `title:"CAPACITY" def:"0"`
	Allocatable float64 `title:"ALLOCATABLE" def:"0"`
	Allocated   float64 `title:"ALLOCATED"`
	InUse       int     `title:"IN USE"`
	Free        int     `title:"FREE"`
	Util        float64 `title:"Util." format:"%"`
	Unhealthy   int     `title:"UNHEALTHY"`
}

type NodeMemoryResource struct {
	Capacity    float64 `title:"CAPACITY" format:"memory" def:"0"`
	Allocatable float64 `title:"ALLOCATABLE" format:"memory"`
	Allocated   float64 `title:"ALLOCATED" format:"memory"`
	Usage       string  `title:"USAGE"`
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
	Mem    *NodeMemoryResource `group:"MEMORY"`
	GPUs   *NodeGPUResource    `group:"GPU" def:"<none>"`
	GPUMem *NodeMemoryResource `group:"GPU MEMORY" def:"<none>"`
}

type ClusterNodesView struct {
	GPUs            int
	UnhealthyGPUs   int
	GPUsInUse       int
	AllocatedGpus   float64
	GPUsOnReadyNode int
}
