package types

type GPU struct {
	IndexID                          string  `title:"GPU"`
	Allocated                        float64 `title:"ALLOCATED"`
	Utilization                      float64 `title:"UTILIZATION" format:"%"`
	Memory                           float64 `title:"MEMORY" format:"memory"`
	MemoryUsage                      float64 `title:"MEMORY USAGE" format:"memory"`
	MemoryUtilization				 float64 `title:"MEMORY UTILIZATION" format:"%"`
	MemoryUsageAndUtilization		 string  `title:"MEMORY USAGE"`
	IdleTime                         float64 `title:"IDLE TIME" format:"time"`
}
