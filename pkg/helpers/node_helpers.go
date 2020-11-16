package helpers

import (
	"github.com/run-ai/runai-cli/pkg/types"
)

func AddNodeGPUsToClusterNodes(cnv *types.ClusterNodesView, status types.NodeStatus, gpu *types.NodeGPUResource) {
	if gpu == nil {
		return
	}
	cnv.GPUs += gpu.Capacity
	cnv.GPUsInUse += gpu.InUse
	cnv.AllocatedGpus += gpu.Allocated
	cnv.UnhealthyGPUs += gpu.Unhealthy
	if status == types.NodeReady {
		cnv.GPUsOnReadyNode += gpu.Capacity
	}
}
