package util

import (
	v1 "k8s.io/api/core/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"


	"strconv"
	"fmt"
)


// the following code copied from top node

// Does the node have unhealthy GPU
func HasUnhealthyGPU(node v1.Node) (unhealthy bool) {

	totalGPU := TotalGpuInNode(node)
	allocatableGPU := AllocatableGpuInNode(node)

	unhealthy = totalGPU > allocatableGPU

	if unhealthy {
		log.Debugf("node: %s, allocated GPUs %s, total GPUs %s is unhealthy", node.Name, strconv.FormatInt(totalGPU, 10),
			strconv.FormatInt(allocatableGPU, 10))
	}

	return unhealthy
}

func IsMasterNode(node v1.Node) bool {
	if _, ok := node.Labels[masterLabelRole]; ok {
		return true
	}

	return false
}



func GetTotalNodeMemory(node *v1.Node) (totalMemory string) {

	valTotal, ok := node.Status.Capacity["memory"]
	if ok {
		return fmt.Sprintf("%dM", valTotal.ScaledValue(resource.Mega))
	}

	return ""
}

