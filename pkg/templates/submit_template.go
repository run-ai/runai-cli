package templates

import (
	yaml "gopkg.in/yaml.v2"
	"os"
)

type SubmitTemplate struct {
	EnvVariables []string `yaml:"environment,omitempty"`
	Volumes []string `yaml:"volumes,omitempty"`
}

func GetSubmitTemplateFromYaml(templateYaml string, expandEnvVariables bool) (*SubmitTemplate, error) {
	if expandEnvVariables {
		templateYaml = os.ExpandEnv(templateYaml)
	}
	var template SubmitTemplate
	err := yaml.Unmarshal([]byte(templateYaml), template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

