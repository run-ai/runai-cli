package templates

import (
	yaml "gopkg.in/yaml.v2"
	"os"
)

type SubmitTemplate struct {
	EnvVariables []string `yaml:"environment,omitempty"`
	Volumes []string `yaml:"volumes,omitempty"`
	AlwaysPullImage *bool `yaml:"always-pull-image,omitempty"`
	Attach *bool `yaml:"attach,omitempty"`
	Cpu string `yaml:"cpu,omitempty"`
	CpuLimit string `yaml:"cpu-limit,omitempty"`
	CreateHomeDir *bool `yaml:"create-home-dir,omitempty"`
	Gpu *float64 `yaml:"gpu,omitempty"`
	HostIpc *bool `yaml:"host-ipc,omitempty"`
	HostNetwork *bool `yaml:"host-network,omitempty"`
	Image string `yaml:"image,omitempty"`
	Interactive *bool `yaml:"interactive,omitempty"`
	LargeShm *bool `yaml:"large-shm,omitempty"`
	LocalImage *bool `yaml:"local-image,omitempty"`
	Memory string `yaml:"memory,omitempty"`
	MemoryLimit string `yaml:"memory-limit,omitempty"`
	NodeType string `yaml:"node-type,omitempty"`
	Ports []string `yaml:"ports,omitempty"`
	PersistentVolumes []string `yaml:"pvcs,omitempty"`
	WorkingDir string `yaml:"working-dir,omitempty"`
	JobNamePrefix string `yaml:"job-name-prefix,omitempty"`
	PreventPrivilegeEscalation *bool `yaml:"prevent-privilege-escalation,omitempty"`
	RunAsCurrentUser *bool `yaml:"run-as-user,omitempty"`

	BackoffLimit *int `yaml:"backofflimit,omitempty"`
	Elastic *bool `yaml:"elastic,omitempty"`
	Parallelism *int `yaml:"parallelism,omitempty"`
	IsPreemptible *bool `yaml:"preemptible,omitempty"`
	ServiceType string `yaml:"service-type,omitempty"`
	IsJupyter *bool `yaml:"jupyter,omitempty"`

	Processes *int `yaml:"processes,omitempty"`
}

func GetSubmitTemplateFromYaml(templateYaml string, expandEnvVariables bool) (*SubmitTemplate, error) {
	if expandEnvVariables {
		templateYaml = os.ExpandEnv(templateYaml)
	}
	var template SubmitTemplate
	err := yaml.Unmarshal([]byte(templateYaml), &template)
	if err != nil {
		return nil, err
	}

	return &template, nil
}

