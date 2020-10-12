package helpers


import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"strconv"

	"github.com/run-ai/runai-cli/pkg/ui"
	"github.com/run-ai/runai-cli/cmd/types"

)

func RenderClusterNodesView(w io.Writer, cnv types.ClusterNodesView) {

	ui.Title(w, "CLUSTER NODES INFO")

	fmt.Fprintf(w, "Allocated Units/Total GPUs In Cluster:\t")
	log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(int64(cnv.GPUs), 10),
		strconv.FormatInt(int64(cnv.AllocatedGPUs), 10))
	var gpuUsage float64 = 0
	if cnv.GPUs > 0 {
		gpuUsage = float64(cnv.AllocatedGPUs) / float64(cnv.GPUs) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		strconv.FormatInt(int64(cnv.AllocatedGPUs), 10),
		strconv.FormatInt(int64(cnv.GPUs), 10),
		int64(gpuUsage),
	)
	if cnv.GPUs != cnv.GPUsOnReadyNode {
		if cnv.GPUsOnReadyNode > 0 {
			gpuUsage = float64(cnv.AllocatedGPUs) / float64(cnv.GPUsOnReadyNode) * 100
		} else {
			gpuUsage = 0
		}
		fmt.Fprintf(w, "Allocated Units/Total GPUs(Active) In Cluster:\t")
		fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
			strconv.FormatInt(int64(cnv.AllocatedGPUs), 10),
			strconv.FormatInt(int64(cnv.GPUsOnReadyNode), 10),
			int64(gpuUsage))
	}

	if float64(cnv.AllocatedGPUs) != cnv.AllocatedFractionGPUs {
		if cnv.GPUsOnReadyNode > 0 {
			gpuUsage = float64(cnv.AllocatedFractionGPUs) / float64(cnv.GPUsOnReadyNode) * 100
		} else {
			gpuUsage = 0
		}
		fmt.Fprintf(w, "Allocated Fractions/Total GPUs In Cluster:\t")
		fmt.Fprintf(w, "%.1f/%s (%d%%)\t\n",
			cnv.AllocatedFractionGPUs,
			strconv.FormatInt(int64(cnv.GPUsOnReadyNode), 10),
			int64(gpuUsage))
	}

	if cnv.UnhealthyGPUs > 0 {
		fmt.Fprintf(w, "Unhealthy/Total GPUs In Cluster:\n")
		var gpuUnhealthyPercentage float64 = 0
		if cnv.GPUs > 0 {
			gpuUnhealthyPercentage = float64(cnv.UnhealthyGPUs) / float64(cnv.GPUs) * 100
		}
		fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
			strconv.FormatInt(int64(cnv.UnhealthyGPUs), 10),
			strconv.FormatInt(int64(cnv.GPUs), 10),
			int64(gpuUnhealthyPercentage))
	}
}

func AddNodeToClusterNodes(cnv *types.ClusterNodesView, status types.NodeStatus, gpu types.NodeGPUResource) {
	cnv.GPUs += gpu.Capacity
	cnv.AllocatedGPUs += gpu.AllocatedUnits
	cnv.AllocatedFractionGPUs += float64(gpu.AllocatedFraction)
	cnv.UnhealthyGPUs += gpu.Unhealthy
	if status == types.NodeReady {
		cnv.GPUsOnReadyNode += gpu.Capacity
	}
}
