package submit

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestMergeEnvironmentVariablesNoConflict(t *testing.T) {
	cliEnvVar := []string{ "user=test-user" }
	templateEnvVar := []string{ "server=127.0.0.1" }

	mergeResult := mergeEnvironmentVariables(&cliEnvVar, &templateEnvVar)

	assert.Equal(t, len(mergeResult), 2)
}

func TestMergeEnvironmentVariablesOverride(t *testing.T) {
	cliEnvVar := []string{ "user=test-user" }
	templateEnvVar := []string{ "user=override" }

	mergeResult := mergeEnvironmentVariables(&cliEnvVar, &templateEnvVar)

	assert.Equal(t, len(mergeResult), 1)
	assert.Equal(t, mergeResult[0], "user=test-user")
}

func TestMergeBoolFlags(t *testing.T) {
	cliBoolFlag := false
	templateBoolFlag := true

	mergeResult := mergeBoolFlags(&cliBoolFlag, &templateBoolFlag)

	assert.Equal(t, *mergeResult, cliBoolFlag)
}

func TestMergeStringFlags(t *testing.T) {
	cliStringFlag := "CliTest"
	templateStringFlag := "TemplateTest"

	mergeResult := mergeStringFlags(cliStringFlag, templateStringFlag)

	assert.Equal(t, mergeResult, cliStringFlag)
}

func TestMergeFloat64Flag(t *testing.T) {
	cliFloat64Flag := 3.14
	templateFloat64Flag := 42.5

	mergeResult := mergeFloat64Flags(&cliFloat64Flag, &templateFloat64Flag)

	assert.Equal(t, *mergeResult, cliFloat64Flag)
}

func TestMergeCommandAndArgsOnlyTemplateCommand(t *testing.T) {
	templateIsCommand := true
	templateExtraArgs := []string{"bash"}
	cliIsCommand := false
	cliExtraArgs := []string{}

	command, args := mergeCommandAndArgs(templateIsCommand, cliIsCommand, templateExtraArgs, cliExtraArgs)

	assert.Equal(t, len(args), 0)
	assert.Equal(t, len(command), 1)
	assert.Equal(t, command[0], "bash")
}

func TestMergeCommandAndArgsTemplateCommandOverrideByCli(t *testing.T) {
	templateIsCommand := true
	templateExtraArgs := []string{"bash"}
	cliIsCommand := true
	cliExtraArgs := []string{"python"}

	command, args := mergeCommandAndArgs(templateIsCommand, cliIsCommand, templateExtraArgs, cliExtraArgs)

	assert.Equal(t, len(args), 0)
	assert.Equal(t, len(command), 1)
	assert.Equal(t, command[0], "python")
}

func TestMergeCommandAndArgsTemplateCommandWithCliArgs(t *testing.T) {
	templateIsCommand := true
	templateExtraArgs := []string{"bash"}
	cliIsCommand := false
	cliExtraArgs := []string{"echo"}

	command, args := mergeCommandAndArgs(templateIsCommand, cliIsCommand, templateExtraArgs, cliExtraArgs)

	assert.Equal(t, len(args), 1)
	assert.Equal(t, len(command), 1)
	assert.Equal(t, command[0], "bash")
	assert.Equal(t, args[0], "echo")
}

func TestMergeCommandAndArgsTemplateArgsOverride(t *testing.T) {
	templateIsCommand := false
	templateExtraArgs := []string{"bash"}
	cliIsCommand := false
	cliExtraArgs := []string{"echo"}

	command, args := mergeCommandAndArgs(templateIsCommand, cliIsCommand, templateExtraArgs, cliExtraArgs)

	assert.Equal(t, len(args), 1)
	assert.Equal(t, len(command), 0)
	assert.Equal(t, args[0], "echo")
}
