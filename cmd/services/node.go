package services

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	t "github.com/run-ai/runai-cli/cmd/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

)


type NodeDescriber struct {
	client  kubernetes.Interface
	allPods []v1.Pod
}

func NewNodeDescriber(client kubernetes.Interface, pods []v1.Pod) *NodeDescriber {
	return &NodeDescriber{
		client:  client,
		allPods: pods,
	}
}

func (d *NodeDescriber) GetAllNodeInfos() ([]t.NodeInfo, error) {
	nodeInfoList := []t.NodeInfo{}

	nodeList, err := d.client.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		return nodeInfoList, err
	}

	for _, node := range nodeList.Items {

		pods := d.GetPodsFromNode(node)
		nodeInfo := t.NewNodeInfo(
			node,
			pods,
		)
		nodeInfoList = append(nodeInfoList, nodeInfo)
	}
	return nodeInfoList, nil
}

func (d *NodeDescriber) GetPodsFromNode(node v1.Node) []v1.Pod {
	pods := []v1.Pod{}
	for _, pod := range d.allPods {
		if pod.Spec.NodeName == node.Name {
			pods = append(pods, pod)
		}
	}

	return pods
}



