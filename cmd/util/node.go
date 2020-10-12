package util

import (
	"strconv"
	"strings"
	"fmt"

	v1 "k8s.io/api/core/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"

)

const (
	LabelNodeRolePrefix = "node-role.kubernetes.io/"	
	MasterLabelRole = "node-role.kubernetes.io/master"
	// NodeLabelRole specifies the role of a node
	NodeLabelRole = "kubernetes.io/role"
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
	if _, ok := node.Labels[MasterLabelRole]; ok {
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

// GetNodeRoles returns the roles of a given node.
// The roles are determined by looking for:
// * a node-role.kubernetes.io/<role>="" label
// * a kubernetes.io/role="<role>" label
func GetNodeRoles(node *v1.Node) []string {
	roles := sets.NewString()
	for k, v := range node.Labels {
		switch {
		case strings.HasPrefix(k, LabelNodeRolePrefix):
			if role := strings.TrimPrefix(k, LabelNodeRolePrefix); len(role) > 0 {
				roles.Insert(role)
			}

		case k == NodeLabelRole && v != "":
			roles.Insert(v)
		}
	}
	return roles.List()
}

func GetNodeInternalAddress(node v1.Node) string {
	if len(node.Status.Addresses) > 0 {
		for _, addr := range node.Status.Addresses {
			if addr.Type == v1.NodeInternalIP {
				return addr.Address
			}
		}
	}
	return "unknown"
}

func IsNodeReady(node v1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}