package types


import (
	"k8s.io/api/core/v1"
)
const (
	NVIDIAGPUResourceName = "nvidia.com/gpu"
	DeprecatedNVIDIAGPUResourceName = "alpha.kubernetes.io/nvidia-gpu"
)


// it can be limited, requested 
type ResourceList struct {
	Cpu int64
	GPUs int64
	Memory int64
	GPUMemory int64
	Storage int64
}


func (rl *ResourceList) Add(rl2 ResourceList) {
	
	rl.Cpu += rl2.Cpu
	rl.GPUs += rl2.GPUs
	rl.Memory += rl2.Memory
	rl.GPUMemory += rl2.GPUMemory
	rl.Storage += rl2.Storage

}


func (ra *ResourceList) AddKubeResourceList(ra2 v1.ResourceList) {

	ra.Cpu += kubeQuantityToInt64(ra2, v1.ResourceCPU)
	ra.GPUs += kubeQuantityToInt64(ra2, NVIDIAGPUResourceName) + kubeQuantityToInt64(ra2, DeprecatedNVIDIAGPUResourceName)
	ra.Memory += kubeQuantityToInt64(ra2, v1.ResourceMemory)
	ra.Storage += kubeQuantityToInt64(ra2, v1.ResourceStorage)
	
}

func kubeQuantityToInt64(rl v1.ResourceList, key v1.ResourceName) int64 {
	num, ok := rl[key]
	if ok {
		return num.Value()
	}
	return 0
}


