package types

import (
	appsv1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodTemplateJob struct {
	ExtraStatus string
	Type        ResourceType
	metav1.ObjectMeta
	Selector *metav1.LabelSelector
	Template v1.PodTemplateSpec
}

func PodTemplateJobFromJob(job batch.Job) *PodTemplateJob {
	extraStatus := ""
	if job.Status.CompletionTime != nil {
		extraStatus = "Completed"
	} else if job.Spec.BackoffLimit != nil && job.Status.Failed >= *job.Spec.BackoffLimit {
		extraStatus = "Failed"
	} else if job.Status.Active == 0 {
		extraStatus = "Pending"
	}

	return &PodTemplateJob{
		ExtraStatus: extraStatus,
		ObjectMeta:  job.ObjectMeta,
		Type:        ResourceTypeJob,
		Template:    job.Spec.Template,
		Selector:    job.Spec.Selector,
	}
}

func PodTemplateJobFromStatefulSet(statefulSet appsv1.StatefulSet) *PodTemplateJob {
	extraStatus := ""
	if statefulSet.Status.Replicas == 0 {
		extraStatus = "Pending"
	}

	return &PodTemplateJob{
		ExtraStatus: extraStatus,
		ObjectMeta:  statefulSet.ObjectMeta,
		Type:        ResourceTypeStatefulSet,
		Template:    statefulSet.Spec.Template,
		Selector:    statefulSet.Spec.Selector,
	}
}

func PodTemplateJobFromReplicaSet(replicaSet appsv1.ReplicaSet) *PodTemplateJob {
	return &PodTemplateJob{
		ObjectMeta: replicaSet.ObjectMeta,
		Type:       ResourceTypeReplicaset,
		Template:   replicaSet.Spec.Template,
		Selector:   replicaSet.Spec.Selector,
	}
}
