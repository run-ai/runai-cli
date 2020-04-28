package commands

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	runaiGPUFraction = "gpu-fraction"
	runaiGPUIndex    = "runai-gpu"
)

func handleRequestedGPUs(clientset kubernetes.Interface, submitArgs *submitRunaiJobArgs) error {
	if submitArgs.GPU == nil {
		return nil
	}

	if float64(int(*submitArgs.GPU)) == *submitArgs.GPU {
		gpu := int(*submitArgs.GPU)
		submitArgs.GPUInt = &gpu

		return nil
	}

	interactiveJobPatch := true
	submitArgs.Interactive = &interactiveJobPatch
	err := validateFractionalGPUTask(submitArgs)
	if err != nil {
		return err
	}

	submitArgs.GPUFraction = fmt.Sprintf("%v", *submitArgs.GPU)

	if *submitArgs.GPU >= 0.3 {
		submitArgs.GPUFractionFixed = fmt.Sprintf("%v", (*submitArgs.GPU)*0.8)
	} else {
		submitArgs.GPUFractionFixed = fmt.Sprintf("%v", (*submitArgs.GPU)*0.7)
	}

	if *submitArgs.GPU >= 0.5 {
		submitArgs.Args = []string{"32", "224"}
	} else {
		submitArgs.Args = []string{"128", "32"}
	}

	// patch for demo
	submitArgs.Image = "gcr.io/run-ai-lab/quickstart-sharing"

	setConfigMapForFractionalGPU(clientset, submitArgs.Name)
	return nil
}

func validateFractionalGPUTask(submitArgs *submitRunaiJobArgs) error {
	if submitArgs.Interactive == nil || *submitArgs.Interactive == false {
		return fmt.Errorf("Jobs that require a fractional number of GPUs must be interactive. Run the job with flag '--interactive'")
	}

	if submitArgs.Elastic != nil && *submitArgs.Elastic == true {
		return fmt.Errorf("Jobs that require a fractional number of GPUs can't be elastic jobs. Run the job without flag '--elastic'")
	}

	if *submitArgs.GPU > 1 {
		return fmt.Errorf("Jobs that require a fractional number of GPUs must require less than 1 GPU")
	}

	return nil
}

func setConfigMapForFractionalGPU(clientset kubernetes.Interface, jobName string) error {
	runaiVisibleDevices := "RUNAI-VISIBLE-DEVICES"
	configMapName := fmt.Sprintf("%v-%v", jobName, "runai-fraction-gpu")
	configMap, err := clientset.CoreV1().ConfigMaps("default").Get(configMapName, metav1.GetOptions{})

	// Map already exists
	if err == nil {
		configMap.Data[runaiVisibleDevices] = ""
		_, err = clientset.CoreV1().ConfigMaps("default").Update(configMap)
		return err
	}

	data := make(map[string]string)
	data[runaiVisibleDevices] = ""
	configMap = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configMapName,
		},
		Data: data,
	}

	_, err = clientset.CoreV1().ConfigMaps("default").Create(configMap)
	return err
}
