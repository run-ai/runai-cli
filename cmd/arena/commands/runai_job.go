package commands

import (
	"fmt"
	"strconv"
	"time"

	cmdTypes "github.com/kubeflow/arena/cmd/arena/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RunaiJob struct {
	*cmdTypes.BasicJobInfo
	trainerType       string
	chiefPod          *v1.Pod
	creationTimestamp metav1.Time
	interactive       bool
	createdByCLI      bool
	serviceUrls       []string
	deleted           bool
	podSpec           v1.PodSpec
	podMetadata       metav1.ObjectMeta
	jobMetadata       metav1.ObjectMeta
	namespace         string
	pods              []v1.Pod
	status            string
}

func NewRunaiJob(pods []v1.Pod, lastCreatedPod *v1.Pod, creationTimestamp metav1.Time, trainingType string, jobName string, createdByCLI bool, serviceUrls []string, deleted bool, podSpec v1.PodSpec, podMetadata metav1.ObjectMeta, jobMetadata metav1.ObjectMeta, namespace string, ownerResource cmdTypes.Resource, status string) *RunaiJob {
	resources := append(cmdTypes.PodResources(pods), ownerResource)
	return &RunaiJob{
		pods:              pods,
		BasicJobInfo:      cmdTypes.NewBasicJobInfo(jobName, resources),
		chiefPod:          lastCreatedPod,
		creationTimestamp: creationTimestamp,
		trainerType:       trainingType,
		createdByCLI:      createdByCLI,
		serviceUrls:       serviceUrls,
		deleted:           deleted,
		podSpec:           podSpec,
		podMetadata:       podMetadata,
		jobMetadata:       jobMetadata,
		namespace:         namespace,
		status:            status,
	}
}

// // Get the chief Pod of the Job.
func (rj *RunaiJob) ChiefPod() *v1.Pod {
	return rj.chiefPod
}

// Get the name of the Training Job
func (rj *RunaiJob) Name() string {
	return rj.BasicJobInfo.Name()
}

// Get the namespace of the Training Job
func (rj *RunaiJob) Namespace() string {
	return rj.namespace
}

// Get all the pods of the Training Job
func (rj *RunaiJob) AllPods() []v1.Pod {
	return rj.pods
}

// Get all the kubernetes resource of the Training Job
func (rj *RunaiJob) Resources() []cmdTypes.Resource {
	return rj.BasicJobInfo.Resources()
}

func (rj *RunaiJob) getStatus() v1.PodPhase {
	return rj.chiefPod.Status.Phase
}

func hasPodReadyCondition(conditions []corev1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func GetStatusColumnFromPodStatus(pod *corev1.Pod) string {
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

// Get the Status of the Job: RUNNING, PENDING,
func (rj *RunaiJob) GetStatus() string {
	if value, exists := rj.jobMetadata.Annotations["unschedulable"]; exists {
		if value == "true" {
			return "Unschedulable"
		}
	}

	if rj.status != "" {
		return rj.status
	}

	if rj.chiefPod == nil {
		return "Unknown"
	}

	return GetStatusColumnFromPodStatus(rj.chiefPod)
}

// Return trainer Type, support MPI, standalone, tensorflow
func (rj *RunaiJob) Trainer() string {
	return rj.trainerType
}

// Get the Job Age
func (rj *RunaiJob) Age() time.Duration {
	if rj.creationTimestamp.IsZero() {
		return 0
	}
	return metav1.Now().Sub(rj.creationTimestamp.Time)
}

// TODO
// Get the Job Duration
func (rj *RunaiJob) Duration() time.Duration {
	if rj.chiefPod == nil {
		return 0
	}

	status := rj.getStatus()
	startTime := rj.StartTime()

	if startTime == nil {
		return 0
	}

	var finishTime metav1.Time = metav1.Now()

	if status == v1.PodSucceeded || status == v1.PodFailed {
		// The transition time of ready will be when the pod finished executing
		for _, condition := range rj.ChiefPod().Status.Conditions {
			if condition.Type == v1.PodReady {
				finishTime = condition.LastTransitionTime
			}
		}
	}

	return finishTime.Sub(startTime.Time)
}

func (rj *RunaiJob) CreatedByCLI() bool {
	return rj.createdByCLI
}

// Get start time
func (rj *RunaiJob) StartTime() *metav1.Time {
	if rj.chiefPod == nil {
		return nil
	}

	pod := rj.ChiefPod()
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodInitialized && condition.Status == v1.ConditionTrue {
			return &condition.LastTransitionTime
		}
	}

	return nil
}

func (rj *RunaiJob) GetPodGroupName() string {
	pod := rj.chiefPod
	if pod == nil {
		if len(rj.jobMetadata.Labels) > 0 {
			return rj.jobMetadata.Labels[PodGroupLabel]
		}
		return ""
	}

	if pod.Spec.SchedulerName != SchedulerName {
		return ""
	}

	if len(pod.Labels) > 0 {
		return pod.Labels[PodGroupLabel]
	}
	return ""
}

// Get Dashboard
func (rj *RunaiJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
	return []string{}, nil
}

// Requested GPU count of the Job
func (rj *RunaiJob) RequestedGPU() float64 {
	requestedGPUs := float64(0)
	for _, pod := range rj.pods {
		gpuFraction, GPUFractionErr := strconv.ParseFloat(pod.Annotations[runaiGPUFraction], 64)
		if GPUFractionErr == nil {
			requestedGPUs += gpuFraction
		}
	}

	if requestedGPUs != 0 {
		return requestedGPUs
	}

	val, ok := rj.podSpec.Containers[0].Resources.Limits[NVIDIAGPUResourceName]
	if !ok {
		return 0
	}

	return float64(val.Value())
}

// Requested GPU count of the Job
func (rj *RunaiJob) AllocatedGPU() float64 {
	if rj.chiefPod == nil {
		return 0
	}

	pod := rj.chiefPod

	if pod.Status.Phase == v1.PodRunning {
		return float64(rj.RequestedGPU())
	}

	return 0
}

// the host ip of the chief pod
func (rj *RunaiJob) HostIPOfChief() string {
	if rj.chiefPod == nil {
		return ""
	}

	// This will hold the node name even if not actually specified on pod spec by the user.
	// Copied from describe function of kubectl.
	// https://github.com/kubernetes/kubectl/blob/a20db94d5b5f052d991eaf29d626fb730b4886b7/pkg/describe/versioned/describe.go

	return rj.ChiefPod().Spec.NodeName
}

// The priority class name of the training job
func (rj *RunaiJob) GetPriorityClass() string {
	return ""
}

func (rj *RunaiJob) Image() string {
	return rj.podSpec.Containers[0].Image
}

func (rj *RunaiJob) Project() string {
	return rj.podMetadata.Labels["project"]
}

func (rj *RunaiJob) User() string {
	return rj.podMetadata.Labels["user"]
}

func (rj *RunaiJob) ServiceURLs() []string {
	return rj.serviceUrls
}
