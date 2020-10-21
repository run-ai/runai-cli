package types

type GPU struct {
	IndexID     string  `title:"GPU"`
	Allocated   float64 `title:"ALLOCATED" format:"%"`
	Memory      float64 `title:"MEMORY" format:"memory"`
	MemoryUsage float64 `title:"MEMORY USAGE" format:"memory"`
	IdleTime    float64 `title:"IDLE TIME" format:"time"`
	Used        float64 `title:"USAGE" format:"%"`
}
