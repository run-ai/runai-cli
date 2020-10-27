package submit

import (
	raUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/templates"
	"strings"
)

func applyTemplateToSubmitRunaijob(templateYaml string, args *submitRunaiJobArgs, extraArgs []string) error {
	template, err := templates.GetSubmitTemplateFromYaml(templateYaml, true)
	if err != nil {
		return err
	}

	*args = mergeTemplateToRunaiSubmitArgs(*args, template, extraArgs)
	return nil
}

func applyTemplateToSubmitMpijob(templateYaml string, args *submitMPIJobArgs, extraArgs []string) error {
	template, err := templates.GetSubmitTemplateFromYaml(templateYaml, true)
	if err != nil {
		return err
	}

	*args = mergeTemplateToMpiSubmitArgs(*args, template, extraArgs)
	return nil
}

func mergeTemplateToSubmitArgs(submitArgs submitArgs, template *templates.SubmitTemplate, extraArgs []string) submitArgs {
	submitArgs.NameParameter = mergeStringFlags(submitArgs.NameParameter, template.Name)
	submitArgs.EnvironmentVariable = mergeEnvironmentVariables(&submitArgs.EnvironmentVariable, &template.EnvVariables)
	submitArgs.Volumes = append(submitArgs.Volumes, template.Volumes...)
	submitArgs.AlwaysPullImage = mergeBoolFlags(submitArgs.AlwaysPullImage, template.AlwaysPullImage)
	submitArgs.Attach = mergeBoolFlags(submitArgs.Attach, template.Attach)
	submitArgs.CPU = mergeStringFlags(submitArgs.CPU, template.Cpu)
	submitArgs.CPULimit = mergeStringFlags(submitArgs.CPULimit, template.CpuLimit)
	submitArgs.CreateHomeDir = mergeBoolFlags(submitArgs.CreateHomeDir, template.CreateHomeDir)
	submitArgs.GPU = mergeFloat64Flags(submitArgs.GPU, template.Gpu)
	submitArgs.HostIPC = mergeBoolFlags(submitArgs.HostIPC, template.HostIpc)
	submitArgs.HostNetwork = mergeBoolFlags(submitArgs.HostNetwork, template.HostNetwork)
	submitArgs.Image = mergeStringFlags(submitArgs.Image, template.Image)
	submitArgs.Interactive = mergeBoolFlags(submitArgs.Interactive, template.Interactive)
	submitArgs.LargeShm = mergeBoolFlags(submitArgs.LargeShm, template.LargeShm)
	submitArgs.LocalImage = mergeBoolFlags(submitArgs.LocalImage, template.LocalImage)
	submitArgs.Memory = mergeStringFlags(submitArgs.Memory, template.Memory)
	submitArgs.MemoryLimit = mergeStringFlags(submitArgs.MemoryLimit, template.MemoryLimit)
	submitArgs.Ports = append(submitArgs.Ports, template.Ports...)
	submitArgs.PersistentVolumes = append(submitArgs.PersistentVolumes, template.PersistentVolumes...)
	submitArgs.WorkingDir = mergeStringFlags(submitArgs.WorkingDir, template.WorkingDir)
	submitArgs.NamePrefix = mergeStringFlags(submitArgs.NamePrefix, template.JobNamePrefix)
	submitArgs.PreventPrivilegeEscalation = mergeBoolFlags(submitArgs.PreventPrivilegeEscalation, template.PreventPrivilegeEscalation)
	submitArgs.RunAsCurrentUser = mergeBoolFlags(submitArgs.RunAsCurrentUser, template.RunAsCurrentUser)
	submitArgs.SpecCommand, submitArgs.SpecArgs = mergeCommandAndArgs(raUtil.IsBoolPTrue(template.IsCommand), raUtil.IsBoolPTrue(submitArgs.Command), template.ExtraArgs, extraArgs)
	return submitArgs
}

func mergeTemplateToRunaiSubmitArgs(submitArgs submitRunaiJobArgs, template *templates.SubmitTemplate, extraArgs []string) submitRunaiJobArgs {
	submitArgs.submitArgs = mergeTemplateToSubmitArgs(submitArgs.submitArgs, template, extraArgs)
	submitArgs.BackoffLimit = mergeIntFlags(submitArgs.BackoffLimit, template.BackoffLimit)
	submitArgs.Elastic = mergeBoolFlags(submitArgs.Elastic, template.Elastic)
	submitArgs.Parallelism = mergeIntFlags(submitArgs.Parallelism, template.Parallelism)
	submitArgs.IsPreemptible = mergeBoolFlags(submitArgs.IsPreemptible, template.IsPreemptible)
	submitArgs.ServiceType = mergeStringFlags(submitArgs.ServiceType, template.ServiceType)
	submitArgs.IsJupyter = mergeBoolFlags(submitArgs.IsJupyter, template.IsJupyter)
	return submitArgs
}

func mergeTemplateToMpiSubmitArgs(submitArgs submitMPIJobArgs, template *templates.SubmitTemplate, extraArgs []string) submitMPIJobArgs {
	submitArgs.submitArgs = mergeTemplateToSubmitArgs(submitArgs.submitArgs, template, extraArgs)
	submitArgs.Processes = mergeIntFlags(submitArgs.Processes, template.Processes)
	return submitArgs
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

func mergeIntFlags(cliFlag, templateFlag *int) *int {
	if cliFlag != nil {
		return cliFlag
	} else if templateFlag != nil {
		return templateFlag
	}
	return nil
}

func mergeCommandAndArgs(templateIsCommand, cliIsCommand bool, templateExtraArgs, cliExtraArgs []string) ([]string, []string) {
	if templateIsCommand && !cliIsCommand {
		return templateExtraArgs, cliExtraArgs
	} else if templateIsCommand && cliIsCommand {
		return cliExtraArgs, []string{}
	} else if !templateIsCommand && cliIsCommand {
		return cliExtraArgs, []string{}
	} else {
		if len(cliExtraArgs) != 0 {
			return []string{}, cliExtraArgs
		}
		return []string{}, templateExtraArgs
	}
}