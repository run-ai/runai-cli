package submit

import (
	"github.com/run-ai/runai-cli/pkg/templates"
	"strings"
)

func applyTemplate(templateYaml string, args *submitArgs) error {
	template, err := templates.GetSubmitTemplateFromYaml(templateYaml, true)
	if err != nil {
		return err
	}

	*args = mergeTemplateToSubmitArgs(*args, template)
	return nil
}

func mergeTemplateToSubmitArgs(args submitArgs, template *templates.SubmitTemplate) submitArgs {
	args.EnvironmentVariable = mergeEnvironmentVariables(&args.EnvironmentVariable, &template.EnvVariables)
	args.Volumes = append(args.Volumes, template.Volumes...)
	args.AlwaysPullImage = mergeBoolFlags(args.AlwaysPullImage, template.AlwaysPullImage)
	args.Attach = mergeBoolFlags(args.Attach, template.Attach)
	args.CPU = mergeStringFlags(args.CPU, template.Cpu)
	args.CPULimit = mergeStringFlags(args.CPULimit, template.CpuLimit)
	args.CreateHomeDir = mergeBoolFlags(args.CreateHomeDir, template.CreateHomeDir)
	args.GPU = mergeFloat64Flags(args.GPU, template.Gpu)
	args.HostIPC = mergeBoolFlags(args.HostIPC, template.HostIpc)
	args.HostNetwork = mergeBoolFlags(args.HostNetwork, template.HostNetwork)
	args.Image = mergeStringFlags(args.Image, template.Image)
	args.Interactive = mergeBoolFlags(args.Interactive, template.Interactive)
	args.LargeShm = mergeBoolFlags(args.LargeShm, template.LargeShm)
	args.LocalImage = mergeBoolFlags(args.LocalImage, template.LocalImage)
	args.Memory = mergeStringFlags(args.Memory, template.Memory)
	args.MemoryLimit = mergeStringFlags(args.MemoryLimit, template.MemoryLimit)
	args.Ports = append(args.Ports, template.Ports...)
	args.PersistentVolumes = append(args.PersistentVolumes, template.PersistentVolumes...)
	args.WorkingDir = mergeStringFlags(args.WorkingDir, template.WorkingDir)

	return args
}

func mergeEnvironmentVariables(cliEnvVars, templateEnvVars *[]string) []string {
	cliEnvVarMap := make(map[string]bool)

	for _, cliVar := range *cliEnvVars {
		maybeKeyVal := strings.Split(cliVar, "=")
		if len(maybeKeyVal) != 2 {
			continue
		}
		key := maybeKeyVal[0]
		cliEnvVarMap[key] = true
	}

	for _, templateVar := range *templateEnvVars {
		maybeKeyVal := strings.Split(templateVar, "=")
		if len(maybeKeyVal) != 2 {
			continue
		}
		key := maybeKeyVal[0]
		if !cliEnvVarMap[key] {
			*cliEnvVars = append(*cliEnvVars, templateVar)
		}
	}

	return *cliEnvVars
}

func mergeBoolFlags(cliFlag, templateFlag *bool) *bool {
	if cliFlag != nil {
		return cliFlag
	} else if templateFlag != nil {
		return templateFlag
	}
	return nil
}

func mergeStringFlags(cliFlag, templateFlag string) string {
	if cliFlag != "" {
		return cliFlag
	} else if templateFlag != "" {
		return templateFlag
	}
	return ""
}

func mergeFloat64Flags(cliFlag, templateFlag *float64) *float64 {
	if cliFlag != nil {
		return cliFlag
	} else if templateFlag != nil {
		return templateFlag
	}
	return nil
}