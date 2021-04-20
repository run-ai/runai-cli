package types

import v1 "k8s.io/api/core/v1"

type NodeStatus string



type NodeCPUResource struct {
	Capacity    int     `title:"CAPACITY" def:"0"`
	Allocatable float64 `title:"ALLOCATABLE"`
	Allocated   float64 `title:"ALLOCATED"`
	Utilization float64 `title:"UTILIZATION" format:"%"`
	Usage       float64 `title:"USAGE"`
}

type NodeGPUResource struct {
	GpuType     string  `title:"TYPE" def:"-"`
	Capacity    int     `title:"CAPACITY" def:"0"`
	Allocatable float64 `title:"ALLOCATABLE" def:"0"`
	Allocated   float64 `title:"ALLOCATED"`
	InUse       int     `title:"IN USE"`
	Free        int     `title:"FREE"`
	Utilization float64 `title:"UTILIZATION" format:"%"`
	Usage       float64 `title:"USAGE"`
	Unhealthy   int     `title:"UNHEALTHY"`
}

type NodeMemoryResource struct {
	Capacity            float64 `title:"CAPACITY" format:"memory" def:"0"`
	Allocatable         float64 `title:"ALLOCATABLE" format:"memory"`
	Allocated           float64 `title:"ALLOCATED" format:"memory"`
	Utilization         float64 `title:"UTILIZATION" format:"%"`
	Usage               float64 `title:"USAGE" format:"memory"`
	UsageAndUtilization string  `title:"USAGE"`
}

type NodeGeneralInfo struct {
	Name      string     `title:"NAME"`
	Status    v1.NodeConditionType `title:"STATUS"`
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
