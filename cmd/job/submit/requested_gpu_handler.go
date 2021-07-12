package submit

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	minGpuMemory            = 100
	GpuMbFactor             = 1000000 // 1024 * 1024
	migDeviceResourcePrefix = "nvidia.com/mig-"
)

var supportedMigDevices = map[string]bool{
	"1g.5gb":  true,
	"2g.10gb": true,
	"3g.20gb": true,
	"4g.20gb": true,
	"7g.40gb": true,
}

func handleRequestedGPUs(submitArgs *submitArgs) error {
	if submitArgs.GPU == nil && submitArgs.GPUMemory == "" && submitArgs.MigDevice == "" {
		return nil
	}

	if (submitArgs.GPU != nil) && (submitArgs.GPUMemory != "") {
		return fmt.Errorf("unexpected to accept both gpu and gpu-memory flag. please use only one of them")
	} else if (submitArgs.GPU != nil) && (submitArgs.MigDevice != "") {
		return fmt.Errorf("unexpected to accept both gpu and mig flag. please use only one of them")
	} else if (submitArgs.GPUMemory != "") && (submitArgs.MigDevice != "") {
		return fmt.Errorf("unexpected to accept both gpu-memory and mig flag. please use only one of them")
	}

	if submitArgs.MigDevice != "" {
		if _, ok := supportedMigDevices[submitArgs.MigDevice]; !ok {
			return fmt.Errorf("unsupported mig device: %v", submitArgs.MigDevice)
		}
		submitArgs.MigDevice = fmt.Sprintf("%v%v", migDeviceResourcePrefix, submitArgs.MigDevice)
		return nil
	}

	if submitArgs.GPUMemory != "" {
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
