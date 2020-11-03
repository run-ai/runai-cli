package types

type PodResourcesStatus struct {
	Limited   ResourceList
	Allocated ResourceList
	Requested ResourceList
	Usage     ResourceList
}
