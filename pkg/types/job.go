package types

import (
	"time"
)

// JobGeneralInfo general information
type JobGeneralInfo struct {
	Name     string        `title:"NAME"`
	Project  string        `title:"PROJECT"`
	User     string        `title:"USER"`
	Type     string        `title:"TYPE"`
	Status   string        `title:"STATUS"`
	Duration time.Duration `title:"DURATION" format:"time"`
	Node     string        `title:"NODE"`
}

// ResourceUsage resource usage
type ResourceUsage struct {
	Usage       float64 `title:"Usage"`
	Utilization float64 `title:"Utiliz."`
}

// MemoryMetrics resource metrics
type MemoryMetrics struct {
	Allocated float64        `title:"Allocated" format:"memory"`
	Usage     *ResourceUsage `title:"Usage" format:"memoryusage"`
}

// CPUMetrics resource metrics
type CPUMetrics struct {
	Allocated float64        `title:"Allocated" format:"cpu"`
	Usage     *ResourceUsage `title:"Usage" format:"cpuusage"`
}

// GPUMetrics resource metrics
type GPUMetrics struct {
	Allocated   float64 `title:"Allocated"`
	Utilization float64 `title:"Util" format:"%"`
}

// JobView is general status of a RunAI/MPI Job
type JobView struct {
	Info   *JobGeneralInfo `group:"GENERAL,flatten"`
	GPUs   *GPUMetrics     `group:"GPU" def:"<none>"`
	GPUMem *MemoryMetrics  `group:"GPU MEMORY" def:"<none>"`
	CPUs   *CPUMetrics     `group:"CPU"`
	Mem    *MemoryMetrics  `group:"CPU MEMORY"`
}
