package helpers

import (
	"strings"

	"github.com/run-ai/runai-cli/pkg/types"
	v1 "k8s.io/api/core/v1"
)

func AddNodeGPUsToClusterNodes(cnv *types.ClusterNodesView, status string, gpu *types.NodeGPUResource) {
	if gpu == nil {
		return
	}
	cnv.GPUs += gpu.Capacity
	cnv.GPUsInUse += gpu.InUse
	cnv.AllocatedGpus += gpu.Allocated
	cnv.UnhealthyGPUs += gpu.Unhealthy
	if strings.Contains(status,  string(v1.NodeReady)) {
		cnv.GPUsOnReadyNode += gpu.Capacity
	}
}
