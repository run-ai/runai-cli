package trainer

import (
	runaiclientset "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned"
	"github.com/run-ai/runai-cli/pkg/client"
	"strconv"
	"time"

	"github.com/run-ai/runai-cli/cmd/constants"
	"github.com/run-ai/runai-cli/cmd/util"
	cmdTypes "github.com/run-ai/runai-cli/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const runaiJobTrainerName = "Runaijob"

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
	parallelism       int32
	completions       int32
	failed            int32
	succeeded         int32
	workloadType      cmdTypes.ResourceType
}

func NewRunaiJob(pods []v1.Pod, lastCreatedPod *v1.Pod, creationTimestamp metav1.Time, trainingType string, jobName string, createdByCLI bool, serviceUrls []string, deleted bool, podSpec v1.PodSpec, podMetadata metav1.ObjectMeta, jobMetadata metav1.ObjectMeta, namespace string, ownerResource cmdTypes.Resource, status string, parallelism, completions, failed, succeeded int32) *RunaiJob {
	workloadType := ownerResource.ResourceType
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
		parallelism:       parallelism,
		completions:       completions,
		failed:            failed,
		succeeded:         succeeded,
		workloadType:      workloadType,
	}
}

func (rj *RunaiJob) TrainerName() string {
	return runaiJobTrainerName
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

// Get the Status of the Job: RUNNING, PENDING,
func (rj *RunaiJob) GetStatus() string {
	return rj.status
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

	startTime := rj.StartTime()

	if startTime == nil {
		return 0
	}

	return rj.FinishTime().Sub(startTime.Time)
}

func (rj *RunaiJob) CreatedByCLI() bool {
	return rj.createdByCLI
}

// Get start time
func (rj *RunaiJob) StartTime() *metav1.Time {
	if rj.parallelism > 1 {
		var earliestStartTime *metav1.Time
		for _, pod := range rj.pods {
			startTimeOfPod := getStartTimeOfPod(&pod)
			if startTimeOfPod != nil && (earliestStartTime == nil || earliestStartTime.After(startTimeOfPod.Time)) {
				earliestStartTime = startTimeOfPod
			}
		}
		return earliestStartTime
	}

	return getStartTimeOfPod(rj.chiefPod)
}

// Get start time
func (rj *RunaiJob) FinishTime() *metav1.Time {
	if rj.parallelism > 1 {
		now := metav1.Now()
		latestEndTimeOfPod := &now
		for _, pod := range rj.pods {
			endTimeOfPod := getEndTimeOfPod(&pod)
			if endTimeOfPod != nil && (latestEndTimeOfPod == nil || latestEndTimeOfPod.Before(endTimeOfPod)) {
				latestEndTimeOfPod = endTimeOfPod
			}
		}
		return latestEndTimeOfPod
	}

	return getEndTimeOfPod(rj.chiefPod)
}

func getStartTimeOfPod(pod *v1.Pod) *metav1.Time {
	if pod == nil {
		return nil
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodInitialized && condition.Status == v1.ConditionTrue {
			return &condition.LastTransitionTime
		}
	}
	return nil
}

func getEndTimeOfPod(pod *v1.Pod) *metav1.Time {
	finishTime := metav1.Now()
	if pod == nil {
		return &finishTime
	}

	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		// The transition time of ready will be when the pod finished executing
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodReady {
				return &condition.LastTransitionTime
			}
		}
	}
	return &finishTime
}

func (rj *RunaiJob) GetPodGroupName() string {
	pod := rj.chiefPod
	if pod == nil {
		if len(rj.jobMetadata.Annotations) > 0 {
			return rj.jobMetadata.Annotations[constants.PodGroupAnnotationForPod]
		}
		return ""
	}

	if pod.Spec.SchedulerName != constants.SchedulerName {
		return ""
	}

	if len(pod.Labels) > 0 {
		return pod.Annotations[constants.PodGroupAnnotationForPod]
	}
	return ""
}

// Get Dashboard
func (rj *RunaiJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
	return []string{}, nil
}

// Requested GPU count of the Job
func (rj *RunaiJob) RequestedGPU() float64 {
	requestedGPUs, ok := util.GetRequestedGPUsPerPodGroup(rj.jobMetadata.Annotations)
	if ok {
		return requestedGPUs
	}

	// backward compatibility
	for _, pod := range rj.pods {
		gpuFraction, GPUFractionErr := strconv.ParseFloat(pod.Annotations[util.RunaiGPUFraction], 64)
		if GPUFractionErr == nil {
			requestedGPUs += gpuFraction
		}
	}

	if requestedGPUs != 0 {
		return requestedGPUs
	}

	val, ok := rj.podSpec.Containers[0].Resources.Limits[util.NVIDIAGPUResourceName]
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

	nodeName, ok := getNodeName(rj.jobMetadata.Annotations)
	if ok {
		return nodeName
	}

	// This will hold the node name even if not actually specified on pod spec by the user.
	// Copied from describe function of kubectl.
	// https://github.com/kubernetes/kubectl/blob/a20db94d5b5f052d991eaf29d626fb730b4886b7/pkg/describe/versioned/describe.go

	return rj.ChiefPod().Spec.NodeName // backward compatibility
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

func (rj *RunaiJob) RunningPods() int32 {
	runningPods, ok := getRunningPods(rj.jobMetadata.Annotations)
	if ok {
		return runningPods
	}

	// backward compatibility
	runningPods = 0
	for _, pod := range rj.pods {
		if pod.Status.Phase == v1.PodRunning {
			runningPods++
		}
	}
	return runningPods
}

func (rj *RunaiJob) PendingPods() int32 {
	pendingPods, ok := getPendingPods(rj.jobMetadata.Annotations)
	if ok {
		return pendingPods
	}

	// backward compatibility
	pendingPods = 0
	for _, pod := range rj.pods {
		if pod.Status.Phase == v1.PodPending {
			pendingPods++
		}
	}
	return pendingPods
}

func (rj *RunaiJob) Completions() int32 {
	return rj.completions
}

func (rj *RunaiJob) Parallelism() int32 {
	return rj.parallelism
}

func (rj *RunaiJob) Succeeded() int32 {
	return rj.succeeded

}

func (rj *RunaiJob) Failed() int32 {
	return rj.failed
}

func (rj *RunaiJob) CurrentRequestedGPUs() float64 {
	totalRequestedGPUs, ok := getCurrentRequestedGPUs(rj.jobMetadata.Annotations)
	if ok {
		return totalRequestedGPUs
	}

	// backward compatibility
	if IsFinishedStatus(rj.GetStatus()) {
		return 0
	}

	if rj.chiefPod == nil {
		return 0
	}

	if rj.chiefPod.Status.Phase != v1.PodRunning && rj.chiefPod.Status.Phase != v1.PodPending {
		return 0
	}
	return rj.RequestedGPU()
}

func (rj *RunaiJob) CurrentAllocatedGPUs() float64 {
	totalRequestedGPUs, ok := getAllocatedRequestedGPUs(rj.jobMetadata.Annotations)
	if ok {
		return totalRequestedGPUs
	}

	// backward compatibility
	if rj.chiefPod == nil {
		return 0
	}

	if rj.chiefPod.Status.Phase != v1.PodRunning {
		return 0
	}
	return rj.RequestedGPU()
}

func (rj *RunaiJob) WorkloadType() string {
	return string(rj.workloadType)
}

func (rj *RunaiJob) TotalRequestedGPUs() float64 {
	totalRequestedGPUs, ok := getTotalAllocatedGPUs(rj.jobMetadata.Annotations)
	if ok {
		return totalRequestedGPUs
	}

	return rj.RequestedGPU() * float64(rj.Completions())
}

func (rj *RunaiJob) Delete(kubeClient *client.Client) error {
	if rj.Trainer() == "Interactive" {
		return kubeClient.GetClientset().AppsV1().StatefulSets(rj.Namespace()).Delete(rj.Name(), &metav1.DeleteOptions{})
	}

	clientset, err := runaiclientset.NewForConfig(kubeClient.GetRestConfig())
	if err != nil {
		return err
	}

	return clientset.RunV1().RunaiJobs(rj.Namespace()).Delete(rj.Name(), metav1.DeleteOptions{})
}
