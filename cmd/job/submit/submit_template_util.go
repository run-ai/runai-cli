package submit

import (
	"fmt"
	log "github.com/golang/glog"
	raUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/templates"
	"strconv"
	"strings"
	"time"
)

func recoverFromMissingFlag(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf(r.(string))
	}
}

func applyTemplateToSubmitRunaijob(template *templates.SubmitTemplate, args *submitRunaiJobArgs, extraArgs []string) (err error) {
	defer recoverFromMissingFlag(&err)

	*args = mergeTemplateToRunaiSubmitArgs(*args, template, extraArgs)
	return nil
}

func applyTemplateToSubmitMpijob(template *templates.SubmitTemplate, args *submitMPIJobArgs, extraArgs []string) (err error) {
	defer (func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(r.(string))
		}
	})()

	*args = mergeTemplateToMpiSubmitArgs(*args, template, extraArgs)
	return nil
}

func mergeTemplateToCommonSubmitArgs(submitArgs submitArgs, template *templates.SubmitTemplate, extraArgs []string) submitArgs {
	submitArgs.NameParameter = applyTemplateFieldForString(submitArgs.NameParameter, template.Name, "name")
	submitArgs.EnvironmentVariable = mergeEnvironmentVariables(&submitArgs.EnvironmentVariable, &template.EnvVariables)
	submitArgs.Volumes = append(submitArgs.Volumes, template.Volumes...)
	submitArgs.AlwaysPullImage = applyTemplateFieldForBool(submitArgs.AlwaysPullImage, template.AlwaysPullImage, "always-pull-image")
	submitArgs.Attach = applyTemplateFieldForBool(submitArgs.Attach, template.Attach, "attach")
	submitArgs.CPU = applyTemplateFieldForString(submitArgs.CPU, template.Cpu, "cpu")
	submitArgs.CPULimit = applyTemplateFieldForString(submitArgs.CPULimit, template.CpuLimit, "cpu-limit")
	submitArgs.CreateHomeDir = applyTemplateFieldForBool(submitArgs.CreateHomeDir, template.CreateHomeDir, "create-home-dir")
	submitArgs.GPU = applyTemplateFieldForFloat64(submitArgs.GPU, template.Gpu, "gpu")
	submitArgs.HostIPC = applyTemplateFieldForBool(submitArgs.HostIPC, template.HostIpc, "host-ipc")
	submitArgs.HostNetwork = applyTemplateFieldForBool(submitArgs.HostNetwork, template.HostNetwork, "host-network")
	submitArgs.Image = applyTemplateFieldForString(submitArgs.Image, template.Image, "image")
	submitArgs.Interactive = applyTemplateFieldForBool(submitArgs.Interactive, template.Interactive, "interactive")
	submitArgs.LargeShm = applyTemplateFieldForBool(submitArgs.LargeShm, template.LargeShm, "large-shm")
	submitArgs.LocalImage = applyTemplateFieldForBool(submitArgs.LocalImage, template.LocalImage, "local-image")
	submitArgs.Memory = applyTemplateFieldForString(submitArgs.Memory, template.Memory, "memory")
	submitArgs.MemoryLimit = applyTemplateFieldForString(submitArgs.MemoryLimit, template.MemoryLimit, "memory-limit")
	submitArgs.Ports = append(submitArgs.Ports, template.Ports...)
	submitArgs.PersistentVolumes = append(submitArgs.PersistentVolumes, template.PersistentVolumes...)
	submitArgs.WorkingDir = applyTemplateFieldForString(submitArgs.WorkingDir, template.WorkingDir, "working-dir")
	submitArgs.NamePrefix = applyTemplateFieldForString(submitArgs.NamePrefix, template.JobNamePrefix, "job-name-prefix")
	submitArgs.PreventPrivilegeEscalation = applyTemplateFieldForBool(submitArgs.PreventPrivilegeEscalation, template.PreventPrivilegeEscalation, "prevent-privilege-escalation")
	submitArgs.RunAsCurrentUser = applyTemplateFieldForBool(submitArgs.RunAsCurrentUser, template.RunAsCurrentUser, "run-as-user")
	submitArgs.Command = applyTemplateFieldForBool(submitArgs.Command, template.IsCommand, "command")
	mergeCommandAndArgs(&submitArgs, template, extraArgs)
	return submitArgs
}

func mergeTemplateToRunaiSubmitArgs(submitArgs submitRunaiJobArgs, template *templates.SubmitTemplate, extraArgs []string) submitRunaiJobArgs {
	submitArgs.submitArgs = mergeTemplateToCommonSubmitArgs(submitArgs.submitArgs, template, extraArgs)
	submitArgs.BackoffLimit = applyTemplateFieldForInt(submitArgs.BackoffLimit, template.BackoffLimit, "backofflimit")
	submitArgs.Elastic = applyTemplateFieldForBool(submitArgs.Elastic, template.Elastic, "elastic")
	submitArgs.Parallelism = applyTemplateFieldForInt(submitArgs.Parallelism, template.Parallelism, "parallelism")
	submitArgs.IsPreemptible = applyTemplateFieldForBool(submitArgs.IsPreemptible, template.IsPreemptible, "preemptible")
	submitArgs.ServiceType = applyTemplateFieldForString(submitArgs.ServiceType, template.ServiceType, "service-type")
	submitArgs.IsJupyter = applyTemplateFieldForBool(submitArgs.IsJupyter, template.IsJupyter, "jupyter")
	submitArgs.TtlAfterFinished = applyTemplateFieldForDuration(submitArgs.TtlAfterFinished, template.TtlAfterFinished, "ttl-after-finish")
	return submitArgs
}

func mergeTemplateToMpiSubmitArgs(submitArgs submitMPIJobArgs, template *templates.SubmitTemplate, extraArgs []string) submitMPIJobArgs {
	submitArgs.submitArgs = mergeTemplateToCommonSubmitArgs(submitArgs.submitArgs, template, extraArgs)
	submitArgs.Processes = applyTemplateFieldForInt(submitArgs.Processes, template.Processes, "processes")
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

func mergeDurationFlags(cliFlag, templateFlag *time.Duration) *time.Duration {
	if cliFlag != nil {
		return cliFlag
	} else if templateFlag != nil {
		return templateFlag
	}
	return nil
}

func mergeExtraArgs(cliExtraArgs, templateExtraArgs []string) []string {
	if len(cliExtraArgs) > 0 {
		return cliExtraArgs
	} else if len(templateExtraArgs) > 0 {
		return templateExtraArgs
	}

	return []string{}
}

func mergeCommandAndArgs(submitArgs *submitArgs, template *templates.SubmitTemplate, extraArgs []string) {
	submitArgs.Command = applyTemplateFieldForBool(submitArgs.Command, template.IsCommand, "command")
	if raUtil.IsBoolPTrue(submitArgs.Command) {
		submitArgs.SpecCommand = mergeExtraArgs(extraArgs, template.ExtraArgs)
		submitArgs.SpecArgs = []string{}
	} else {
		submitArgs.SpecCommand = []string{}
		submitArgs.SpecArgs = mergeExtraArgs(extraArgs, template.ExtraArgs)
	}
}


func applyTemplateFieldForFloat64(cliFlag *float64, templateField *templates.TemplateField, fieldName string) *float64 {
	var value *float64
	required := false
	var templateFlag *float64
	if templateField != nil {
		required = raUtil.IsBoolPTrue(templateField.Required)
		templateFieldValue, err := strconv.ParseFloat(templateField.Value, 64)
		if err == nil && templateField.Value != "" {
			log.Info(fmt.Sprintf("could not parse %s flag from template", fieldName))
			templateFlag = &templateFieldValue
		}
	}

	value = mergeFloat64Flags(cliFlag, templateFlag)
	validateValueIsNotRequiredAndNil(value == nil, required, fieldName)
	return value
}

func applyTemplateFieldForInt(cliFlag *int, templateField *templates.TemplateField, fieldName string) *int {
	var value *int
	required := false
	var templateFlag *int
	if templateField != nil {
		required = raUtil.IsBoolPTrue(templateField.Required)
		templateFieldValue, err := strconv.Atoi(templateField.Value)
		if err == nil && templateField.Value != "" {
			log.Info(fmt.Sprintf("could not parse %s flag from template", fieldName))
			templateFlag = &templateFieldValue
		}
	}

	value = mergeIntFlags(cliFlag, templateFlag)
	validateValueIsNotRequiredAndNil(value == nil, required, fieldName)
	return value
}

func applyTemplateFieldForBool(cliFlag *bool, templateField *templates.TemplateField, fieldName string) *bool {
	var value *bool
	required := false
	var templateFlag *bool
	if templateField != nil {
		required = raUtil.IsBoolPTrue(templateField.Required)
		templateFieldValue, err := strconv.ParseBool(templateField.Value)
		if err == nil && templateField.Value != "" {
			log.Info(fmt.Sprintf("could not parse %s flag from template", fieldName))
			templateFlag = &templateFieldValue
		}
	}

	value = mergeBoolFlags(cliFlag, templateFlag)
	validateValueIsNotRequiredAndNil(value == nil, required, fieldName)
	return value
}

func applyTemplateFieldForDuration(cliFlag *time.Duration, templateField *templates.TemplateField, fieldName string) *time.Duration {
	var value *time.Duration
	required := false
	var templateFlag *time.Duration
	if templateField != nil {
		required = raUtil.IsBoolPTrue(templateField.Required)
		templateFieldValue, err := time.ParseDuration(templateField.Value)
		if err == nil && templateField.Value != "" {
			log.Info(fmt.Sprintf("could not parse %s flag from template", fieldName))
			templateFlag = &templateFieldValue
		}
	}

	value = mergeDurationFlags(cliFlag, templateFlag)
	validateValueIsNotRequiredAndNil(value == nil, required, fieldName)
	return value
}

func applyTemplateFieldForString(cliFlag string, templateField *templates.TemplateField, fieldName string) string {
	var value string
	required := false
	if templateField != nil {
		required = raUtil.IsBoolPTrue(templateField.Required)
		value = mergeStringFlags(cliFlag, templateField.Value)
	} else {
		value = mergeStringFlags(cliFlag, "")
	}

	if value == "" && required {
		panic(fmt.Sprintf("Field %s is required", fieldName))
	}
	return value
}

func validateValueIsNotRequiredAndNil(valueIsNil, required bool, fieldName string) {
	if valueIsNil && required {
		panic(fmt.Sprintf("Field %s is required", fieldName))
	}
}