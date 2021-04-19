package constants

const (
	RUNAI_QUEUE_LABEL = "runai/queue"
	RUNAI_NS_PROJECT_PREFIX = "runai-"
)

// Same statuses appear in the scheduler - update both if needed
var Status = struct {
	Running   string
	Pending   string
	Succeeded string
	Deleted   string
	Failed    string
	TimedOut  string
	Preempted string
	Unknown   string
}{
	Running:   "Running",
	Pending:   "Pending",
	Succeeded: "Succeeded",
	Deleted:   "Deleted",
	Failed:    "Failed",
	TimedOut:  "TimedOut",
	Preempted: "Preempted",
	Unknown:   "Unknown",
}

// todo organize

const (

	PodGroupAnnotationForPod = "pod-group-name"

	CHART_PKG_LOC = "CHARTREPO"
	SchedulerName = "runai-scheduler"

	masterLabelRole = "node-role.kubernetes.io/master"

	gangSchdName = "kube-batchd"

	// labelNodeRolePrefix is a label prefix for node roles
	// It's copied over to here until it's merged in core: https://github.com/kubernetes/kubernetes/pull/39112
	labelNodeRolePrefix = "node-role.kubernetes.io/"

	WorkloadCalculatedStatus     = "runai-calculated-status"
	WorkloadRunningPods          = "runai-running-pods"
	WorkloadPendingPods          = "runai-pending-pods"
	WorkloadUsedNodes            = "runai-used-nodes"
	AliyunENIAnnotation          = "k8s.aliyun.com/eni"
)

