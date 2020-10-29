package types

type GPU struct {
	IndexID     string  `title:"GPU"`
	Allocated   float64 `title:"ALLOCATED" format:"%"`
	Util        float64 `title:"UTIL." format:"%"`
	Memory      float64 `title:"MEMORY" format:"memory"`
	MemoryUsage float64 `title:"MEMORY USAGE" format:"memory"`
	IdleTime    float64 `title:"IDLE TIME" format:"time"`
}
