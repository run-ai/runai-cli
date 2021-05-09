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

// JobCPUUsage job cpu usage
type JobCPUUsage struct {
	Requested   float64 `title:"REQUESTED"`
	Utilization float64 `title:"UTILIZATION" format:"%"`
}

// JobMemoryUsage job memory usage
type JobMemoryUsage struct {
	Allocated   int64   `title:"ALLOCATED" format:"memory"`
	Used        int64   `title:"USED" format:"memory"`
	Utilization float64 `title:"UTILIZATION" format:"%"`
}

// JobGPUUsage job gpu usage
type JobGPUUsage struct {
	Requested   float64 `title:"REQUESTED"`
	Utilization float64 `title:"UTILIZATION" format:"%"`
}

// ResourceUsage resource usage
type ResourceUsage struct {
	Usage       float64 `title:"USAGE"`
	Utilization float64 `title:"UTILIZ."`
}

// MemoryMetrics resource metrics
type MemoryMetrics struct {
	Allocated float64        `title:"ALLOCATED" format:"memory"`
	Usage     *ResourceUsage `title:"Usage" format:"memoryusage"`
}

// CPUMetrics resource metrics
type CPUMetrics struct {
	Allocated float64        `title:"ALLOCATED"`
	Usage     *ResourceUsage `title:"Usage" format:"cpuusage"`
}

// JobView is general status of a RunAI/MPI Job
type JobView struct {
	Info   *JobGeneralInfo `group:"GENERAL,flatten"`
	GPUs   *CPUMetrics     `group:"GPU" def:"<none>"`
	GPUMem *MemoryMetrics  `group:"GPU MEMORY" def:"<none>"`
	CPUs   *CPUMetrics     `group:"CPU"`
	Mem    *MemoryMetrics  `group:"CPU MEMORY"`
}
