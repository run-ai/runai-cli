package submit

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	minGpuMemory = 100
	GpuMbFactor  = 1000000 // 1024 * 1024
)

func handleRequestedGPUs(submitArgs *submitArgs) error {
	if submitArgs.GPU == nil && submitArgs.GPUMemory == "" {
		return nil
	} else if submitArgs.GPU != nil && submitArgs.GPUMemory != "" {
		return fmt.Errorf("unexpected to accept both gpu and gpu-memory flag. please use only one of them")
	} else if submitArgs.GPU == nil {
		memoryQuantity, err := resource.ParseQuantity(submitArgs.GPUMemory)
		if err != nil {
			return err
		}

		memoryInMib := memoryQuantity.Value() / GpuMbFactor //From bytes to mib
		if memoryInMib < minGpuMemory {
			return fmt.Errorf("gpu memory must be greater than 100Mb")
		}

		submitArgs.GPUMemory = fmt.Sprintf("%d", memoryInMib)
		return nil
	}

	if float64(int(*submitArgs.GPU)) == *submitArgs.GPU {
		gpu := int(*submitArgs.GPU)
		submitArgs.GPUInt = &gpu
		return nil
	}

	submitArgs.GPUFraction = fmt.Sprintf("%g", *submitArgs.GPU)
	return nil
}
