package types

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"strconv"

	"github.com/run-ai/runai-cli/pkg/ui"
)

type NodeStatus string

const (
	NodeReady    NodeStatus = "ready"
	NodeNotReady NodeStatus = "notReady"
)

type NodeCPUResource struct {
	Capacity    int     `title:"CAPACITY"`
	Allocatable float64 `title:"ALLOCATABLE"`
	Requested   float64 `title:"REQUESTED"`
	// Limit float64				`title:"Limit"`
	Usage float64 `title:"USAGE" format:"%"`
}

type NodeGPUResource struct {
	Capacity          int     `title:"CAPACITY"`
	Allocatable       float64 `title:"ALLOCATABLE"`
	Unhealthy         int     `title:"UNHEALTHY"`
	Allocated         int     `title:"ALLOCATED UNITS"`
	AllocatedFraction float64 `title:"ALLOCATED FRACTION"`
	Usage             float64 `title:"USAGE" format:"%"`
}

type NodeMemoryResource struct {
	Capacity    float64 `title:"CAPACITY" format:"memory"`
	Allocatable float64 `title:"ALLOCATABLE" format:"memory"`
	Requested   float64 `title:"REQUESTED" format:"memory"`
	// Limit float64				`title:"Limit"`
	Usage float64 `title:"USAGE" format:"memory"`
}

type NodeGeneralInfo struct {
	Name      string     `title:"NAME"`
	Status    NodeStatus `title:"STATUS"`
	IPAddress string     `title:"IP Address"`
	Role      string     `title:"ROLE" def:"<none>"`
}

type NodeView struct {
	Info   NodeGeneralInfo    `group:"GENERAL,flatten"`
	CPUs   NodeCPUResource    `group:"CPU"`
	GPUs   NodeGPUResource    `group:"GPU"`
	Mem    NodeMemoryResource `group:"MEMORY"`
	GPUMem NodeMemoryResource `group:"GPU MEMORY"`
}

type ClusterNodesView struct {
	GPUs            int
	UnhealthyGPUs   int
	AllocatedGPUs   float64
	GPUsOnReadyNode int
}

func (cnv *ClusterNodesView) Render(w io.Writer) {

	ui.Title(w, "CLUSTER NODES INFO")

	fmt.Fprintf(w, "Allocated/Total GPUs In Cluster:\t")
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
			gpuUsage = cnv.AllocatedGPUs / float64(cnv.GPUsOnReadyNode) * 100
		} else {
			gpuUsage = 0
		}
		fmt.Fprintf(w, "Allocated/Total GPUs(Active) In Cluster:\t")
		fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
			strconv.FormatInt(int64(cnv.AllocatedGPUs), 10),
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

func (cnv *ClusterNodesView) AddNode(status NodeStatus, gpu NodeGPUResource) {
	cnv.GPUs += gpu.Capacity
	cnv.AllocatedGPUs += float64(gpu.Allocated)
	cnv.UnhealthyGPUs += gpu.Unhealthy
	if status == NodeReady {
		cnv.GPUsOnReadyNode += gpu.Capacity
	}
}
