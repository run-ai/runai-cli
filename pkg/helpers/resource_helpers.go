package helpers

import (
	"github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/types"
	v1 "k8s.io/api/core/v1"
)

func AddToResourceList(rl *types.ResourceList, rl2 types.ResourceList) {
	rl.CPUs += rl2.CPUs
	rl.GPUs += rl2.GPUs
	rl.Memory += rl2.Memory
	rl.GPUMemory += rl2.GPUMemory
	rl.Storage += rl2.Storage
}

func AddKubeResourceListToResourceList(rl *types.ResourceList, krl v1.ResourceList) {
	rl.CPUs += kubeQuantityToMilliFloat64(krl, v1.ResourceCPU)
	rl.GPUs += kubeQuantityToFloat64(krl, util.NVIDIAGPUResourceName)
	rl.Memory += kubeQuantityToFloat64(krl, v1.ResourceMemory)
	rl.Storage += kubeQuantityToFloat64(krl, v1.ResourceStorage)
}
