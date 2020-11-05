package templates

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Template struct {
	Name        string
	Description string
	Values      string
	IsDefault   bool
}

type Templates struct {
	clientset kubernetes.Interface
}

const (
	runaiNamespace         = "runai"
	runaiConfigLabel       = "runai/template"
	runaiDefaultAnnotation = "runai/admin"
)

func NewTemplates(clientset kubernetes.Interface) Templates {
	return Templates{
		clientset: clientset,
	}
}

func (cg *Templates) ListTemplates() ([]Template, error) {
	configsList, err := cg.clientset.CoreV1().ConfigMaps(runaiNamespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=true", runaiConfigLabel),
	})

	if err != nil {
		return []Template{}, err
	}

	log.Debugf("Found %d templates", len(configsList.Items))

	var clusterConfigs []Template

	for _, config := range configsList.Items {
		clusterConfig := Template{}

		if config.Annotations != nil {
			clusterConfig.IsDefault = config.Annotations[runaiDefaultAnnotation] == "true"
		}

		clusterConfig.Name = config.Data["name"]
		clusterConfig.Description = config.Data["description"]
		clusterConfig.Values = config.Data["values"]
		clusterConfigs = append(clusterConfigs, clusterConfig)
	}

	return clusterConfigs, nil
}

func (cg *Templates) GetTemplate(name string) (*Template, error) {
	configs, err := cg.ListTemplates()
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.Name == name {
			return &config, err
		}
	}

	return nil, nil
}

func (cg *Templates) GetDefaultTemplate() (*Template, error) {
	configs, err := cg.ListTemplates()
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.IsDefault {
			return &config, err
		}
	}

	return nil, nil
}
