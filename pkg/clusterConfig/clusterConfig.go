package clusterConfig

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ClusterConfig struct {
	EnforceRunAsUser                  bool `yaml:"enforceRunAsUser"`
	EnforcePreventPrivilegeEscalation bool `yaml:"enforcePreventPrivilegeEscalation"`
}

const (
	runaiNamespace          = "runai"
	runaiClusterConfigLabel = "runai/cluster-config"
	runaiConfigKey          = "config"
)

func GetClusterConfig(clientset kubernetes.Interface) (*ClusterConfig, error) {
	configsList, err := clientset.CoreV1().ConfigMaps(runaiNamespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=true", runaiClusterConfigLabel),
	})

	if err != nil {
		return nil, err
	}

	defaultConfig := ClusterConfig{}

	if configsList == nil || len(configsList.Items) == 0 {
		return &defaultConfig, nil
	}

	configMap := configsList.Items[0]

	if configString, ok := configMap.Data[runaiConfigKey]; ok {
		err := yaml.Unmarshal([]byte(configString), &defaultConfig)
		if err != nil {
			return &defaultConfig, err
		} else {
			return &defaultConfig, nil
		}
	} else {
		return &defaultConfig, fmt.Errorf("error reading cluster configuration, could not find %s key on configmap data", runaiConfigKey)
	}
}
