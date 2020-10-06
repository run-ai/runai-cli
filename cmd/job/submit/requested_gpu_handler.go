package submit

import (
	"fmt"
)

const (
	runaiGPUFraction = "gpu-fraction"
	runaiGPUIndex    = "runai-gpu"
)

func handleRequestedGPUs(submitArgs *submitArgs) {
	if submitArgs.GPU == nil {
		return
	}

	if float64(int(*submitArgs.GPU)) == *submitArgs.GPU {
		gpu := int(*submitArgs.GPU)
		submitArgs.GPUInt = &gpu
		return
	}

	submitArgs.GPUFraction = fmt.Sprintf("%g", *submitArgs.GPU)
	return
}
