package submit

import (
	"fmt"
	"github.com/run-ai/researcher-service/server/pkg/runai/api"
	"github.com/run-ai/researcher-service/server/pkg/schema"
	raUtil "github.com/run-ai/runai-cli/cmd/util"
	"os"
	"reflect"
	"time"
)

func enforce(value interface{}, field schema.SettingsEnforcer, name string) interface{} {
	if reflect.ValueOf(field).IsNil() {
		return value
	}
	val, err := field.Enforce(value, name)
	if err != nil {
		fmt.Printf("Usage: %s\n", err.Error())
		os.Exit(1)
	}
	return val
}

func enforceImagePolicy(isAlwaysPull *bool, imagePullPolicy *schema.StringField, paramName string) *bool {
	if imagePullPolicy == nil {
		return isAlwaysPull
	}
	policy := "IfNotPresent"
	if isAlwaysPull != nil && *isAlwaysPull {
		policy = "Always"
	}

	policy = enforce(policy, imagePullPolicy, paramName).(string)

	return schema.BoolRef(policy == "Always")
}

func enforceDuration(duration *time.Duration, durationPolicy *schema.StringField, paramName string) *time.Duration {
	if durationPolicy == nil {
		return duration
	}
	var durationAsStr string
	if duration != nil {
		durationAsStr = fmt.Sprintf("%vs", duration.Seconds())
	}

	durationAsStr = enforce(durationAsStr, durationPolicy, paramName).(string)
	if durationAsStr == "" {
		return nil
	}

	resultDuration, err := time.ParseDuration(durationAsStr)
	if err != nil {
		fmt.Print("Invalid duration '%s' provided for %s: %s", durationAsStr, paramName, err.Error())
	}
	return &resultDuration
}

func recoverFromMissingFlag(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf(r.(string))
	}
}

func applyTemplateToSubmitRunaijob(args *submitRunaiJobArgs, jobSettings *api.JobSettings, extraArgs []string) (err error) {
	defer recoverFromMissingFlag(&err)

	*args = mergeTemplateToRunaiSubmitArgs(*args, jobSettings, extraArgs)
	return nil
}

func applyTemplateToSubmitMpijob(args *submitMPIJobArgs, jobSettings *api.JobSettings, extraArgs []string) (err error) {
	defer recoverFromMissingFlag(&err)

	*args = mergeTemplateToMpiSubmitArgs(*args, jobSettings, extraArgs)
	return nil
}

func mergeTemplateToCommonSubmitArgs(submitArgs submitArgs, jobSettings *api.JobSettings, extraArgs []string) submitArgs {
	submitArgs.NameParameter = enforce(submitArgs.NameParameter, jobSettings.Fields.Name, "name").(string)
	submitArgs.EnvironmentVariable = enforce(submitArgs.EnvironmentVariable, jobSettings.Fields.Environment, "environment").([]string)
	submitArgs.AlwaysPullImage = enforceImagePolicy(submitArgs.AlwaysPullImage, jobSettings.Fields.ImagePullPolicy, "always-pull-image")
	submitArgs.CPU = enforce(submitArgs.CPU, jobSettings.Fields.Cpu, "cpu").(string)
	submitArgs.CPULimit = enforce(submitArgs.CPULimit, jobSettings.Fields.CpuLimit, "cpu-limit").(string)
	submitArgs.CreateHomeDir = enforce(submitArgs.CreateHomeDir, jobSettings.Fields.CreateHomeDir, "create-home-dir").(*bool)
	submitArgs.GPU = enforce(submitArgs.GPU, jobSettings.Fields.Gpu, "gpu").(*float64)
	submitArgs.HostIPC = enforce(submitArgs.HostIPC, jobSettings.Fields.HostIpc, "host-ipc").(*bool)
	submitArgs.HostNetwork = enforce(submitArgs.HostNetwork, jobSettings.Fields.HostNetwork, "host-network").(*bool)
	submitArgs.Image = enforce(submitArgs.Image, jobSettings.Fields.Image, "image").(string)
	submitArgs.Image = enforce(submitArgs.Image, jobSettings.Fields.Image, "image").(string)
	submitArgs.LargeShm = enforce(submitArgs.LargeShm, jobSettings.Fields.LargeShm, "large-shm").(*bool)
	submitArgs.Memory = enforce(submitArgs.Memory, jobSettings.Fields.Memory, "memory").(string)
	submitArgs.MemoryLimit = enforce(submitArgs.MemoryLimit, jobSettings.Fields.MemoryLimit, "memory-limit").(string)
	submitArgs.Ports = enforce(submitArgs.Ports, jobSettings.Fields.Ports, "ports").([]string)
	submitArgs.PersistentVolumes = enforce(submitArgs.PersistentVolumes, jobSettings.Fields.Pvc, "pvc").([]string)
	submitArgs.WorkingDir = enforce(submitArgs.WorkingDir, jobSettings.Fields.WorkingDir, "working-dir").(string)
	submitArgs.NamePrefix = enforce(submitArgs.NamePrefix, jobSettings.Fields.NamePrefix, "job-name-prefix").(string)
	submitArgs.PreventPrivilegeEscalation = enforce(submitArgs.PreventPrivilegeEscalation, jobSettings.Fields.PreventPrivilegeEscalation, "prevent-privilege-escalation").(*bool)
	submitArgs.Command = enforce(submitArgs.Command, jobSettings.Fields.Command, "command").(*bool)
	mergeGitSync(&submitArgs, jobSettings)
	mergeCommandAndArgs(&submitArgs, jobSettings, extraArgs)
	return submitArgs
}

func mergeGitSync(submitArgs *submitArgs, jobSettings *api.JobSettings) {
	if submitArgs.GitSync == nil {
		submitArgs.GitSync = NewGitSync()
	}
}

func mergeTemplateToRunaiSubmitArgs(submitArgs submitRunaiJobArgs, jobSettings *api.JobSettings, extraArgs []string) submitRunaiJobArgs {
	submitArgs.submitArgs = mergeTemplateToCommonSubmitArgs(submitArgs.submitArgs, jobSettings, extraArgs)
	submitArgs.BackoffLimit = enforce(submitArgs.BackoffLimit, jobSettings.Fields.BackoffLimit, "backofflimit").(*int)
	submitArgs.Elastic = enforce(submitArgs.Elastic, jobSettings.Fields.Elastic, "elastic").(*bool)
	submitArgs.Parallelism = enforce(submitArgs.Parallelism, jobSettings.Fields.Parallelism, "parallelism").(*int)
	submitArgs.IsPreemptible = enforce(submitArgs.IsPreemptible, jobSettings.Fields.Preemptible, "preemptible").(*bool)
	submitArgs.ServiceType = enforce(submitArgs.ServiceType, jobSettings.Fields.ServiceType, "service-type").(string)
	submitArgs.IsJupyter = enforce(submitArgs.IsJupyter, jobSettings.Fields.Jupyter, "jupyter").(*bool)
	submitArgs.TtlAfterFinished = enforceDuration(submitArgs.TtlAfterFinished, jobSettings.Fields.TtlSecondsAfterFinished, "ttl-after-finish")
	return submitArgs
}

func mergeTemplateToMpiSubmitArgs(submitArgs submitMPIJobArgs, jobSettings *api.JobSettings, extraArgs []string) submitMPIJobArgs {
	submitArgs.submitArgs = mergeTemplateToCommonSubmitArgs(submitArgs.submitArgs, jobSettings, extraArgs)
	submitArgs.Processes = enforce(submitArgs.Processes, jobSettings.Fields.MpiProcs, "processes").(*int)
	return submitArgs
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

func mergeCommandAndArgs(submitArgs *submitArgs, jobSettings *api.JobSettings, extraArgs []string) {
	submitArgs.Command = enforce(submitArgs.Command, jobSettings.Fields.Command, "command").(*bool)
	if raUtil.IsBoolPTrue(submitArgs.Command) {
		submitArgs.SpecCommand = enforce(submitArgs.SpecCommand, jobSettings.Fields.Arguments, "command").([]string)
		submitArgs.SpecArgs = []string{}
	} else {
		submitArgs.SpecCommand = []string{}
		submitArgs.SpecArgs = enforce(submitArgs.SpecArgs, jobSettings.Fields.Arguments, "arguments").([]string)
	}
}

func validateValueIsNotRequiredAndNil(valueIsNil, required bool, fieldName string) {
	if valueIsNil && required {
		panic(fmt.Sprintf("the flag %s is mandatory.", fieldName))
	}
}
