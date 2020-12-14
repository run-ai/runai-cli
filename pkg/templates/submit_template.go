package templates

import (
	yaml "gopkg.in/yaml.v2"
	"os"
)

type TemplateField struct {
	Required *bool  `yaml:"required,omitempty"`
	Value    string `yaml:"value,omitempty"`
}

type TemplateListField struct {
	Required *bool    `yaml:"required,omitempty"`
	Value    []string `yaml:"value,omitempty"`
}

type SubmitTemplate struct {
	Name                       *TemplateField   `yaml:"name,omitempty"`
	EnvVariables               []string         `yaml:"environments,omitempty"`
	Volumes                    []string         `yaml:"volumes,omitempty"`
	AlwaysPullImage            *TemplateField   `yaml:"always-pull-image,omitempty"`
	Attach                     *TemplateField   `yaml:"attach,omitempty"`
	Cpu                        *TemplateField   `yaml:"cpu,omitempty"`
	CpuLimit                   *TemplateField   `yaml:"cpu-limit,omitempty"`
	CreateHomeDir              *TemplateField   `yaml:"create-home-dir,omitempty"`
	Gpu                        *TemplateField   `yaml:"gpu,omitempty"`
	HostIpc                    *TemplateField   `yaml:"host-ipc,omitempty"`
	HostNetwork                *TemplateField   `yaml:"host-network,omitempty"`
	Image                      *TemplateField   `yaml:"image,omitempty"`
	Interactive                *TemplateField   `yaml:"interactive,omitempty"`
	LargeShm                   *TemplateField   `yaml:"large-shm,omitempty"`
	LocalImage                 *TemplateField   `yaml:"local-image,omitempty"`
	Memory                     *TemplateField   `yaml:"memory,omitempty"`
	MemoryLimit                *TemplateField   `yaml:"memory-limit,omitempty"`
	NodeType                   *TemplateField   `yaml:"node-type,omitempty"`
	Ports                      []string         `yaml:"ports,omitempty"`
	PersistentVolumes          []string         `yaml:"pvcs,omitempty"`
	WorkingDir                 *TemplateField   `yaml:"working-dir,omitempty"`
	JobNamePrefix              *TemplateField   `yaml:"job-name-prefix,omitempty"`
	PreventPrivilegeEscalation *TemplateField   `yaml:"prevent-privilege-escalation,omitempty"`
	RunAsCurrentUser           *TemplateField   `yaml:"run-as-user,omitempty"`
	ExtraArgs                  []string         `yaml:"extra-args,omitempty"`
	IsCommand                  *TemplateField   `yaml:"command,omitempty"`
	GitSync                    *GitSyncTemplate `yaml:"git-sync,omitempty"`

	BackoffLimit     *TemplateField `yaml:"backofflimit,omitempty"`
	Elastic          *TemplateField `yaml:"elastic,omitempty"`
	Parallelism      *TemplateField `yaml:"parallelism,omitempty"`
	IsPreemptible    *TemplateField `yaml:"preemptible,omitempty"`
	ServiceType      *TemplateField `yaml:"service-type,omitempty"`
	IsJupyter        *TemplateField `yaml:"jupyter,omitempty"`
	TtlAfterFinished *TemplateField `yaml:"ttl-after-finish,omitempty"`

	Processes *TemplateField `yaml:"processes,omitempty"`
}

type GitSyncTemplate struct {
	Repository *TemplateField `yaml:"source,omitempty"`
	Branch     *TemplateField `yaml:"branch,omitempty"`
	Revision   *TemplateField `yaml:"rev,omitempty"`
	Username   *TemplateField `yaml:"username,omitempty"`
	Password   *TemplateField `yaml:"password,omitempty"`
	Image      *TemplateField `yaml:"image,omitempty"`
	Directory  *TemplateField `yaml:"target,omitempty"`
}

func GetSubmitTemplateFromYaml(templateYaml string) (*SubmitTemplate, error) {
	templateYaml = os.ExpandEnv(templateYaml)
	var template SubmitTemplate
	err := yaml.Unmarshal([]byte(templateYaml), &template)
	if err != nil {
		return nil, err
	}

	return &template, nil
}
