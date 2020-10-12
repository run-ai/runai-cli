package helpers

import (
	"github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/types"
	v1 "k8s.io/api/core/v1"
)

func GetPodResourceStatus(pod v1.Pod) types.PodResourcesStatus {

	prs := types.PodResourcesStatus{}

	for _, container := range pod.Spec.Containers {
		AddKubeResourceListToResourceList(&prs.Requested, container.Resources.Requests)
		AddKubeResourceListToResourceList(&prs.Limited, container.Resources.Limits)
	}
	prs.Allocated.GPUs = util.GpuInActivePod(pod)

	return prs
}

func AddToPodResourcesStatus(prs *types.PodResourcesStatus, prs2 types.PodResourcesStatus) {
	AddToResourceList(&prs.Limited, prs2.Limited)
	AddToResourceList(&prs.Allocated, prs2.Allocated)
	AddToResourceList(&prs.Requested, prs2.Requested)
	AddToResourceList(&prs.Usage, prs2.Usage)
}
