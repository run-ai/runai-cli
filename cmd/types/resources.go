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
	CPUs float64
	GPUs float64
	Memory float64
	GPUMemory float64
	Storage float64
}


func (rl *ResourceList) Add(rl2 ResourceList) {
	
	rl.CPUs += rl2.CPUs
	rl.GPUs += rl2.GPUs
	rl.Memory += rl2.Memory
	rl.GPUMemory += rl2.GPUMemory
	rl.Storage += rl2.Storage

}


func (ra *ResourceList) AddKubeResourceList(ra2 v1.ResourceList) {

	ra.CPUs += kubeQuantityToFloat64M(ra2, v1.ResourceCPU)
	ra.GPUs += kubeQuantityToFloat64(ra2, NVIDIAGPUResourceName) + kubeQuantityToFloat64(ra2, DeprecatedNVIDIAGPUResourceName)
	ra.Memory += kubeQuantityToFloat64(ra2, v1.ResourceMemory)
	ra.Storage += kubeQuantityToFloat64(ra2, v1.ResourceStorage)

}

func kubeQuantityToFloat64(rl v1.ResourceList, key v1.ResourceName) float64 {
	num, ok := rl[key]
	if ok {
		return float64(num.Value())
	}
	return 0
}
func kubeQuantityToFloat64M(rl v1.ResourceList, key v1.ResourceName) float64 {
	num, ok := rl[key]
	if ok {
		return float64(num.MilliValue())
	}
	return 0
}
