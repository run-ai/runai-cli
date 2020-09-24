package types


import (
	"fmt"
	"io"
	log "github.com/sirupsen/logrus"
	"strconv"

)



type NodeStatus string

const (
	NodeReady NodeStatus = "ready"
	NodeNotReady NodeStatus = "notReady"
)


type NodeCPUResource struct {
	Capacity int64				`title:"CAPACITY"`
	Allocatable int64			`title:"ALLOCATABLE"`
	Allocated int64				`title:"ALLOCATED"`
	Limited int64				`title:"LIMITED"`
	Requested int64				`title:"REQUESTED"`
	Usage int64					`title:"USAGE"`
}

type NodeGPUResource struct {
	Capacity int64					`title:"CAPACITY"`
	Allocatable int64				`title:"ALLOCATABLE"`
	Unhealthy int64					`title:"UNHEALTHY"`
	Allocated int64					`title:"ALLOCATED UNITS"`
	AllocatedFraction int64  		`title:"ALLOCATED FRACTION"`
	Limited int64					`title:"LIMITED"`
	Requested int64					`title:"REQUESTED"`
	Usage int64						`title:"USAGE"`
}

type NodeMemoryResource struct {
	Capacity int64				`title:"CAPACITY" format:"memory"`
	Allocatable int64			`title:"ALLOCATABLE" format:"memory"`
	Allocated int64				`title:"ALLOCATED" format:"memory"`
	Requested int64				`title:"REQUESTED" format:"memory"`
	Limited int64				`title:"LIMITED" format:"memory"`
	Usage int64					`title:"USAGE" format:"memory"`
}

type NodeStorageResource struct {
	Capacity int64				`title:"CAPACITY" format:"memory"`
	Allocatable int64			`title:"ALLOCATABLE" format:"memory"`
	Allocated int64				`title:"ALLOCATED" format:"memory"`
	Requested int64				`title:"REQUESTED" format:"memory"`
	Limited int64				`title:"LIMITED" format:"memory"`
	Usage int64					`title:"USAGE" format:"memory"`
}

type NodeGeneralInfo struct {
	Name string 					`title:"NAME"`
	IPAddress string 				`title:"IP Address"`
	Role string						`title:"ROLE" def:"<none>"`
	Status NodeStatus				`title:"STATUS"`
}

type NodeView struct {
	Info NodeGeneralInfo            `group:"info" title:"-"`
	CPUs NodeCPUResource			`group:"CPU"`
	GPUs NodeGPUResource			`group:"GPU"`
	Mem NodeMemoryResource			`group:"GPUs MEMORY"`
	GPUMem NodeMemoryResource		`group:"GPUs MEMORY"`
	// todo
	Storage NodeStorageResource     `group:"STORAGE"`
}

type ClusterNodesView struct {
	GPUs            	int64
	UnhealthyGPUs   	int64
	AllocatedGPUs   	int64
	GPUsOnReadyNode 	int64
}


func (cnv *ClusterNodesView) Render(w io.Writer) {

	fmt.Fprintf(w, "-----------------------------------------------------------------------------------------\n")
	
	fmt.Fprintf(w, "Allocated/Total GPUs In Cluster:\t")
	log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(cnv.GPUs, 10),
		strconv.FormatInt(cnv.AllocatedGPUs, 10))
	var gpuUsage float64 = 0
	if cnv.GPUs > 0 {
		gpuUsage = float64(cnv.AllocatedGPUs) / float64(cnv.GPUs) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		strconv.FormatInt(cnv.AllocatedGPUs, 10),
		strconv.FormatInt(cnv.GPUs, 10),
		int64(gpuUsage),
	)
	if cnv.GPUs != cnv.GPUsOnReadyNode {
		if cnv.GPUsOnReadyNode > 0 {
			gpuUsage = float64(cnv.AllocatedGPUs) / float64(cnv.GPUsOnReadyNode) * 100
		} else {
			gpuUsage = 0
		}
		fmt.Fprintf(w, "Allocated/Total GPUs(Active) In Cluster:\t")
		fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
			strconv.FormatInt(cnv.AllocatedGPUs, 10),
			strconv.FormatInt(cnv.GPUsOnReadyNode, 10),
			int64(gpuUsage))
	}

	if cnv.UnhealthyGPUs > 0 {
		fmt.Fprintf(w, "Unhealthy/Total GPUs In Cluster:\n")
		var gpuUnhealthyPercentage float64 = 0
		if cnv.GPUs > 0 {
			gpuUnhealthyPercentage = float64(cnv.UnhealthyGPUs) / float64(cnv.GPUs) * 100
		}
		fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
			strconv.FormatInt(cnv.UnhealthyGPUs, 10),
			strconv.FormatInt(cnv.GPUs, 10),
			int64(gpuUnhealthyPercentage))
	}
}

func (cnv *ClusterNodesView) AddNode(status NodeStatus, gpu NodeGPUResource) {
	cnv.GPUs += gpu.Capacity
	cnv.AllocatedGPUs += gpu.Allocated
	cnv.UnhealthyGPUs += gpu.Unhealthy
	if status == NodeReady{
		cnv.GPUsOnReadyNode += gpu.Capacity
	}
}