package trainer

import (
	"github.com/run-ai/runai-cli/cmd/constants"
	"github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/config"
	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/util/kubectl"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// copy from cmd/common
func GetAllTrainingTypes(kubeClient *client.Client) []string {
	trainers := NewTrainers(kubeClient)
	trainerNames := []string{}
	for _, trainer := range trainers {
		trainerNames = append(trainerNames, trainer.Type())
	}

	return trainerNames
}

// TODO: This method currently take the status from both scheduler, workload and pod's status - The statuses logic should be calculated in 1 place in the future and the logic shouldn't remain as it is today.
func getTrainingStatus(trainingAnnotations map[string]string, chiefPod *v1.Pod, workloadStatus string) string {
	statusValue, statusExists := trainingAnnotations[constants.WorkloadCalculatedStatus]
	// We start by checking finished statuses to identify statuses such as Preempted and TimedOut
	if statusExists && IsFinishedStatus(statusValue) {
		return statusValue
	}

	// Backward compatibility
	if unschedulableValue, isUnschedulableExists := trainingAnnotations["unschedulable"]; isUnschedulableExists {
		if unschedulableValue == "true" {
			return "Unschedulable"
		}
	}

	// We set the status according to the workload before reading the annotation.
	// We do this due to a case where the scheduler wasn't running, the job was completed, pod was deleted - in this case the job annotation will show running.
	// Also there can be a case where the scheduler didn't set the annotation yet but the job already exists.
	if workloadStatus != "" {
		return workloadStatus
	}

	if statusExists {
		return statusValue
	}

	if chiefPod == nil {
		return constants.Status.Unknown
	}

	// Backward compatibility
	return getStatusColumnFromPodStatus(chiefPod)
}

func IsFinishedStatus(status string) bool {
	return status == constants.Status.Succeeded || status == constants.Status.Failed || status == constants.Status.Deleted || status == constants.Status.Preempted || status == constants.Status.TimedOut
}

func getStatusColumnFromPodStatus(pod *corev1.Pod) string {
	// This logic is copied from logic in kubectl
	// Please see https://github.com/kubernetes/kubernetes/blob/a82d71c37621043382c77d00e6e8d47dfb66562e/pkg/printers/internalversion/printers.go#L705
	// for more details

	restarts := 0
	readyContainers := 0

	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}

	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		restarts = 0
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			restarts += int(container.RestartCount)
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
				readyContainers++
			}
		}

		// change pod status back to "Running" if there is at least one container still reporting as "Running" status
		if reason == "Completed" && hasRunning {
			if hasPodReadyCondition(pod.Status.Conditions) {
				reason = "Running"
			} else {
				reason = "NotReady"
			}
		}
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		reason = "Unknown"
	} else if pod.DeletionTimestamp != nil {
		reason = "Terminating"
	}

	return reason
}

func hasPodReadyCondition(conditions []corev1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func getNodeName(trainingAnnotations map[string]string) (string, bool) {
	if len(trainingAnnotations[constants.WorkloadUsedNodes]) > 0 {
		return trainingAnnotations[constants.WorkloadUsedNodes], true
	}
	return "", false
}

func getRunningPods(trainingAnnotations map[string]string) (int32, bool) {
	if len(trainingAnnotations[constants.WorkloadRunningPods]) > 0 {
		runningPods, err := strconv.ParseInt(trainingAnnotations[constants.WorkloadRunningPods], 10, 32)
		if err == nil {
			return int32(runningPods), true
		}
	}
	return 0, false
}

func getPendingPods(trainingAnnotations map[string]string) (int32, bool) {
	if len(trainingAnnotations[constants.WorkloadPendingPods]) > 0 {
		runningPods, err := strconv.ParseInt(trainingAnnotations[constants.WorkloadPendingPods], 10, 32)
		if err == nil {
			return int32(runningPods), true
		}
	}
	return 0, false
}

func getCurrentRequestedGPUs(trainingAnnotations map[string]string) (float64, bool) {
	if len(trainingAnnotations[util.WorkloadCurrentRequestedGPUs]) > 0 {
		totalAllocatedGPUs, err := strconv.ParseFloat(trainingAnnotations[util.WorkloadCurrentRequestedGPUs], 64)
		if err == nil {
			return totalAllocatedGPUs, true
		}
	}
	return 0, false
}

func getCurrentRequestedGPUsMemory(trainingAnnotations map[string]string) (int64, bool) {
	if len(trainingAnnotations[util.WorkloadCurrentRequestedGPUsMemory]) > 0 {
		totalAllocatedGPUs, err := strconv.ParseInt(trainingAnnotations[util.WorkloadCurrentRequestedGPUsMemory], 10, 64)
		if err == nil {
			return totalAllocatedGPUs, true
		}
	}
	return 0, false
}

func getAllocatedRequestedGPUs(trainingAnnotations map[string]string) (float64, bool) {
	if len(trainingAnnotations[util.WorkloadCurrentAllocatedGPUs]) > 0 {
		currentAllocated, err := strconv.ParseFloat(trainingAnnotations[util.WorkloadCurrentAllocatedGPUs], 64)
		if err == nil {
			return currentAllocated, true
		}
	}
	return 0, false
}

func getAllocatedGpusMemory(trainingAnnotations map[string]string) uint64 {
	if len(trainingAnnotations[util.WorkloadCurrentAllocatedGPUsMemory]) > 0 {
		currentAllocated, err := strconv.ParseUint(trainingAnnotations[util.WorkloadCurrentAllocatedGPUsMemory], 10, 64)
		if err == nil {
			return currentAllocated
		}
	}
	return 0
}

func getTotalAllocatedGPUs(trainingAnnotations map[string]string) (float64, bool) {
	if len(trainingAnnotations[util.WorkloadTotalRequestedGPUs]) > 0 {
		totalAllocatedGPUs, err := strconv.ParseFloat(trainingAnnotations[util.WorkloadTotalRequestedGPUs], 64)
		if err == nil {
			return totalAllocatedGPUs, true
		}
	}
	return 0, false
}

func getCliCommand(trainingAnnotations map[string]string) string {
	return trainingAnnotations[util.CliCommand]
}

func getTotalRequestedGPUsMemory(trainingAnnotations map[string]string) uint64 {
	if len(trainingAnnotations[util.WorkloadTotalRequestedGPUsMemory]) > 0 {
		totalRequestedGpusMemory, err := strconv.ParseUint(trainingAnnotations[util.WorkloadTotalRequestedGPUsMemory], 10, 64)
		if err == nil {
			return totalRequestedGpusMemory
		}
	}
	return 0
}

func IsKnownTrainingType(trainingType string) bool {
	for _, knownType := range KnownTrainingTypes {
		if trainingType == knownType {
			return true
		}
	}

	return false
}

/*
* search the training job with name and training type
 */
func SearchTrainingJob(kubeClient *client.Client, jobName string, trainingType string, namespaceInfo types.NamespaceInfo) (job TrainingJob, err error) {
	if len(trainingType) > 0 {
		if IsKnownTrainingType(trainingType) {
			job, err = getTrainingJobByType(kubeClient, jobName, namespaceInfo.Namespace, trainingType)
			if err != nil {
				if isTrainingConfigExist(jobName, trainingType, namespaceInfo.Namespace) {
					log.Warningf("Failed to get the training job %s, but the trainer config is found, please clean it by using '%s delete %s --type %s'.",
						jobName,
						config.CLIName,
						jobName,
						trainingType)
				}
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("%s is unknown training type, please choose a known type from %v",
				trainingType,
				KnownTrainingTypes)
		}
	} else {
		jobs, errorGetByName := getTrainingJobsByName(kubeClient, jobName, namespaceInfo)
		if errorGetByName != nil {
			traningTypes, err := GetTrainingTypes(jobName, namespaceInfo.Namespace, kubeClient.GetClientset())
			if err != nil {
				return nil, err
			}
			if len(traningTypes) > 0 {
				log.Warningf("Failed to get the training job %s, but the trainer config is found, please clean it by using '%s delete %s'.",
					jobName,
					config.CLIName,
					jobName)
			}
			return nil, errorGetByName
		}

		if len(jobs) > 1 {
			return nil, fmt.Errorf("There are more than 1 training jobs with the same name %s, please check it with `%s list | grep %s`",
				jobName,
				config.CLIName,
				jobName)
		} else {
			job = jobs[0]
		}
	}
	return job, nil
}

func getTrainingJobByType(kubeClient *client.Client, name, namespace, trainingType string) (job TrainingJob, err error) {
	// trainers := NewTrainers(client, )

	trainers := NewTrainers(kubeClient)
	for _, trainer := range trainers {
		if trainer.Type() == trainingType {
			return trainer.GetTrainingJob(name, namespace)
		} else {
			log.Debugf("the job %s with type %s in namespace %s is not expected type %v",
				name,
				trainer.Type(),
				namespace,
				trainingType)
		}
	}

	return nil, fmt.Errorf("Failed to find the training job %s in namespace %s", name, namespace)
}

func getTrainingJobsByName(kubeClient *client.Client, name string, namespaceInfo types.NamespaceInfo) (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}
	trainers := NewTrainers(kubeClient)
	for _, trainer := range trainers {
		if trainer.IsSupported(name, namespaceInfo.Namespace) {
			job, err := trainer.GetTrainingJob(name, namespaceInfo.Namespace)
			if err != nil {
				return nil, err
			}
			jobs = append(jobs, job)
		} else {
			log.Debugf("the job %s in namespace %s is not supported by %v", name, namespaceInfo.Namespace, trainer.Type())
		}
	}

	if len(jobs) == 0 {
		log.Debugf("Failed to find the training job %s in namespace %s", name, namespaceInfo.Namespace)
		configMap, err := kubeClient.GetClientset().CoreV1().ConfigMaps(namespaceInfo.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err == nil {
			return nil, fmt.Errorf("The job %s is invalid. Please delete it", configMap.Name)
		}
		return nil, util.GetJobDoesNotExistsInNamespaceError(name, namespaceInfo)
	}

	return jobs, nil
}

/**
*  check if the training config exist
 */
func isTrainingConfigExist(name, trainingType, namespace string) bool {
	configName := fmt.Sprintf("%s-%s", name, trainingType)
	return kubectl.CheckAppConfigMap(configName, namespace)
}

/*
* get App Configs by name, which is created by arena
 */
func GetTrainingTypes(name, namespace string, clientset kubernetes.Interface) (cms []string, err error) {
	configMaps, err := clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return []string{}, err
	}
	cms = []string{}
	for _, trainingType := range KnownTrainingTypes {
		configName := fmt.Sprintf("%s-%s", name, trainingType)
		for _, configMap := range configMaps.Items {
			if configName == configMap.Name {
				cms = append(cms, trainingType)
			}
		}
	}

	return cms, nil
}

func GetGpuMemoryStringFromMemoryCount(memory int64) string {
	gpuMemoryInBytes := memory * GpuMbFactor
	quantity := resource.NewQuantity(gpuMemoryInBytes, resource.DecimalSI)
	return fmt.Sprintf("%v", quantity)
}
