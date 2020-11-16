package util

import (
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	LabelNodeRolePrefix = "node-role.kubernetes.io/"
	// NodeLabelRole specifies the role of a node
	NodeLabelRole = "kubernetes.io/role"
)

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
