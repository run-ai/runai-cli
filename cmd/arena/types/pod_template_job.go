package types

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodTemplateJob struct {
	Type ResourceType
	metav1.ObjectMeta
	Selector *metav1.LabelSelector
	Template v1.PodTemplateSpec
}

func GetPodTemplateJob(objectMeta metav1.ObjectMeta, Template v1.PodTemplateSpec, selector *metav1.LabelSelector, resourceType ResourceType) *PodTemplateJob {
	return &PodTemplateJob{
		ObjectMeta: objectMeta,
		Type:       resourceType,
		Template:   Template,
		Selector:   selector,
	}
}
