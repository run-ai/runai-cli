package helpers

import (
	"github.com/run-ai/runai-cli/pkg/types"
	v1 "k8s.io/api/core/v1"
)

func AddNodeGPUsToClusterNodes(cnv *types.ClusterNodesView, status v1.NodeConditionType, gpu *types.NodeGPUResource) {
	if gpu == nil {
		return
	}
	cnv.GPUs += gpu.Capacity
	cnv.GPUsInUse += gpu.InUse
	cnv.AllocatedGpus += gpu.Allocated
	cnv.UnhealthyGPUs += gpu.Unhealthy
	if status == v1.NodeReady {
		cnv.GPUsOnReadyNode += gpu.Capacity
	}
}
