package types

import (
	"github.com/run-ai/runai-cli/cmd/constants"
	runaijobv1 "github.com/run-ai/runai-cli/cmd/mpi/api/runaijob/v1"
	appsv1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodTemplateJob struct {
	ExtraStatus string // This field is used for backward compatibility, where the scheduler didn't set the real status
	Type        ResourceType
	metav1.ObjectMeta
	Selector    *metav1.LabelSelector
	Template    v1.PodTemplateSpec
	Parallelism int32
	Completions int32
	Failed      int32
	Succeeded   int32
}

func PodTemplateJobFromJob(job batch.Job) *PodTemplateJob {
	extraStatus := ""

	// We don't set the status to succeeded here because of a bug that exists in k8s' job controller, where a pod can be running but the job turns to completed with no reason:
	// this was fixed in: https://github.com/kubernetes/kubernetes/pull/88440 but was not merged to all releases yet.
	for _, event := range job.Status.Conditions {
		if event.Type == "Failed" && event.Status == "True" {
			extraStatus = constants.Status.Failed
			break
		}
	}
	if job.Status.CompletionTime == nil && job.Status.Active == 0 {
		extraStatus = constants.Status.Pending
	}

	parallelism := int32(1)
	if job.Spec.Parallelism != nil {
		parallelism = *job.Spec.Parallelism
	}

	completions := int32(1)
	if job.Spec.Parallelism != nil {
		completions = *job.Spec.Completions
	}
	return &PodTemplateJob{
		ExtraStatus: extraStatus,
		ObjectMeta:  job.ObjectMeta,
		Type:        ResourceTypeJob,
		Template:    job.Spec.Template,
		Selector:    job.Spec.Selector,
		Parallelism: parallelism,
		Completions: completions,
		Failed:      job.Status.Failed,
		Succeeded:   job.Status.Succeeded,
	}
}

func PodTemplateJobFromStatefulSet(statefulSet appsv1.StatefulSet) *PodTemplateJob {
	extraStatus := ""
	if statefulSet.Status.Replicas == 0 {
		extraStatus = constants.Status.Pending
	}

	return &PodTemplateJob{
		ExtraStatus: extraStatus,
		ObjectMeta:  statefulSet.ObjectMeta,
		Type:        ResourceTypeStatefulSet,
		Template:    statefulSet.Spec.Template,
		Selector:    statefulSet.Spec.Selector,
		Parallelism: 1,
		Completions: 1,
		Failed:      0,
		Succeeded:   0,
	}
}

func PodTemplateJobFromReplicaSet(replicaSet appsv1.ReplicaSet) *PodTemplateJob {
	return &PodTemplateJob{
		ObjectMeta:  replicaSet.ObjectMeta,
		Type:        ResourceTypeReplicaset,
		Template:    replicaSet.Spec.Template,
		Selector:    replicaSet.Spec.Selector,
		Parallelism: 1,
		Completions: 1,
		Failed:      0,
		Succeeded:   0,
	}
}

func PodTemplateJobFromRunaiJob(runaiJob runaijobv1.RunaiJob) *PodTemplateJob {
	extraStatus := ""

	// We don't set the status to succeeded here because of a bug that exists in k8s' job controller, where a pod can be running but the job turns to completed with no reason:
	// this was fixed in: https://github.com/kubernetes/kubernetes/pull/88440 but was not merged to all releases yet.
	for _, event := range runaiJob.Status.Conditions {
		if event.Type == "Failed" && event.Status == "True" {
			extraStatus = constants.Status.Failed
			break
		}
	}

	parallelism := int32(1)
	if runaiJob.Spec.Parallelism != nil {
		parallelism = *runaiJob.Spec.Parallelism
	}

	completions := int32(1)
	if runaiJob.Spec.Parallelism != nil && runaiJob.Spec.Completions != nil {
		completions = *runaiJob.Spec.Completions
	}

	selector := &metav1.LabelSelector{}
	if runaiJob.Spec.Selector != nil {
		selector = runaiJob.Spec.Selector
	}
	return &PodTemplateJob{
		ExtraStatus: extraStatus,
		ObjectMeta:  runaiJob.ObjectMeta,
		Type:        ResourceTypeRunaiJob,
		Template:    runaiJob.Spec.Template,
		Selector:    selector,
		Parallelism: parallelism,
		Completions: completions,
		Failed:      runaiJob.Status.Failed,
		Succeeded:   runaiJob.Status.Succeeded,
	}
}
