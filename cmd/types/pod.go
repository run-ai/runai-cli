package types


import (
	v1 "k8s.io/api/core/v1"
)

type PodResourcesStatus struct {
	Limited ResourceList
	Allocated ResourceList
	Requested ResourceList
	Usage ResourceList
}


func getPodResourceStatus(pod v1.Pod) PodResourcesStatus {

	prs := PodResourcesStatus {}

	for _, container := range pod.Spec.Containers {
		prs.Requested.AddKubeResourceList( container.Resources.Requests)
		prs.Limited.AddKubeResourceList( container.Resources.Limits )
		// prs.Allocated
		// prs.Usage
	}

	return prs

}

func (prs *PodResourcesStatus) Add(prs2 PodResourcesStatus) {
	prs.Limited.Add(prs2.Limited)
	prs.Allocated.Add(prs2.Allocated)
	prs.Requested.Add(prs2.Requested)
	prs.Usage.Add(prs2.Usage)
	
}