package node

import (
	"fmt"

	"github.com/run-ai/runai-cli/cmd/trainer"
	"github.com/run-ai/runai-cli/pkg/client"
	nodeService "github.com/run-ai/runai-cli/pkg/services/node"

)

var (
	defaultHiddenFields = []string{
		"Mem.Allocatable",
		"CPUs.Allocatable",
		"GPUs.Allocatable",
		"GPUs.InUse",
		"GPUMem.Allocatable",
		"GPUMem.Requested",
	}
)

func getNodeInfos() (*[]nodeService.NodeInfo, error) {
		kubeClient, err := client.GetClient()
		if err != nil {
			return nil, err
		}
		clientset := kubeClient.GetClientset()
		allPods, err := trainer.AcquireAllActivePods(clientset)
		if err != nil {
			return nil, err
		}
		nd := nodeService.NewNodeDescriber(clientset, allPods)
		nodeInfos, warning, err := nd.GetAllNodeInfos()
		if err != nil {
			return nil, err
		} else if len(warning) > 0 {
			fmt.Println(warning)
		}

		return &nodeInfos, nil
}