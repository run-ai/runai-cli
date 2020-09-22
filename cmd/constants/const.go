package constants

const (
	RUNAI_QUEUE_LABEL = "runai/queue"
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
