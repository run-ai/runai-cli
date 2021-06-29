package templates

import (
	"context"
	"fmt"
	"github.com/run-ai/runai-cli/pkg/config"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Template struct {
	Name        string
	Description string
	Values      string
	IsAdmin     bool
}

type Templates struct {
	clientset kubernetes.Interface
}

const (
	runaiNamespace    = "runai"
	runaiConfigLabel  = "runai/template"
	adminTemplateName = "template-admin"
)

func NewTemplates(clientset kubernetes.Interface) Templates {
	return Templates{
		clientset: clientset,
	}
}

func (cg *Templates) ListTemplates() ([]Template, error) {
	configsList, err := cg.clientset.CoreV1().ConfigMaps(runaiNamespace).List(context.TODO(), metav1.ListOptions{
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
			clusterConfig.IsAdmin = config.Name == adminTemplateName
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
			return &config, nil
		}
	}

	return nil, fmt.Errorf("could not find runai template %s. Please run '%s template list'", name, config.CLIName)
}

func (cg *Templates) GetDefaultTemplate() (*Template, error) {
	configs, err := cg.ListTemplates()
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.IsAdmin {
			return &config, err
		}
	}

	return nil, nil
}
