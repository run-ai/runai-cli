package submit

import (
	"fmt"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestArgsAlignmentNoCommand(t *testing.T) {
	acceptedArgs := []string{"runai", "submit", "-i", "test.io/lab-test", "-g", "1"}

	alignedArgs := AlignArgsPreParsing(acceptedArgs)

	for i, _ := range acceptedArgs {
		if acceptedArgs[i] != alignedArgs[i] {
			t.Errorf("Argument %s changed to %s and the command was not suppose to change", acceptedArgs[i], alignedArgs[i])
		}
	}
}

func TestArgsAlignmentOldCommand(t *testing.T) {
	acceptedArgs := []string{"runai", "submit", "-i", "test.io/lab-test", "-g", "1", "--command", "bash"}

	alignedArgs := AlignArgsPreParsing(acceptedArgs)

	if alignedArgs[6] != fmt.Sprintf("%s%s", dashArg, oldCommandFlag) {
		t.Errorf("The command %s was suppose to change to %s, instead got %s", fmt.Sprintf("%s%s", dashArg, commandFlag), fmt.Sprintf("%s%s", dashArg, oldCommandFlag), alignedArgs[6])
	}
}

func TestArgsAlignmentNewCommand(t *testing.T) {
	acceptedArgs := []string{"runai", "submit", "-i", "test.io/lab-test", "-g", "1", "--command", "--", "bash"}

	alignedArgs := AlignArgsPreParsing(acceptedArgs)

	if alignedArgs[6] != fmt.Sprintf("%s%s", dashArg, commandFlag) {
		t.Errorf("The command %s was suppose to stay to %s, instead got %s", fmt.Sprintf("%s%s", dashArg, commandFlag), fmt.Sprintf("%s%s", dashArg, commandFlag), alignedArgs[6])
	}
}

func TestArgsAlignmentNotSubmit(t *testing.T) {
	acceptedArgs := []string{"runai", "list", "-i", "test.io/lab-test", "-g", "1", "--command", "bash"}

	alignedArgs := AlignArgsPreParsing(acceptedArgs)

	if alignedArgs[6] != fmt.Sprintf("%s%s", dashArg, commandFlag) {
		t.Errorf("The command %s was suppose to stay to %s, instead got %s", fmt.Sprintf("%s%s", dashArg, commandFlag), fmt.Sprintf("%s%s", dashArg, commandFlag), alignedArgs[6])
	}
}

func TestGetSpecCommandAndArgsBackwardCompatibility(t *testing.T) {
	positionalArgs := []string{}
	commandArgs := []string{"bash", "-c"}
	argsArgs := []string{"echo", "test"}
	isCommand := false
	argsLenAtDash := -1

	extraArgs, isCommand := convertOldCommandArgsFlags(argsLenAtDash, positionalArgs, commandArgs, argsArgs, isCommand)

	assert.Equal(t, extraArgs[0], "bash")
	assert.Equal(t, extraArgs[1], "-c")
	assert.Equal(t, extraArgs[2], "echo")
	assert.Equal(t, extraArgs[3], "test")
	assert.Equal(t, isCommand, true)
}

func TestGetSpecCommandAndArgsBackwardCompatibilityOnlyArgs(t *testing.T) {
	commandArgs := []string{}
	positionalArgs := []string{}
	argsArgs := []string{"echo", "test"}
	isCommand := false
	argsLenAtDash := -1

	extraArgs, isCommand := convertOldCommandArgsFlags(argsLenAtDash, positionalArgs, commandArgs, argsArgs, isCommand)

	assert.Equal(t, extraArgs[0], "echo")
	assert.Equal(t, extraArgs[1], "test")
	assert.Equal(t, isCommand, false)
}

func TestGetSpecCommandAndArgsNewCommandAsArgs(t *testing.T) {
	positionalArgs := []string{"sleep", "60"}
	commandArgs := []string{}
	argsArgs := []string{}
	isCommand := false
	argsLenAtDash := 0

	extraArgs, command := convertOldCommandArgsFlags(argsLenAtDash, positionalArgs, commandArgs, argsArgs, isCommand)

	assert.Equal(t, len(extraArgs), 2)
	assert.Equal(t, extraArgs[0], "sleep")
	assert.Equal(t, extraArgs[1], "60")
	assert.Equal(t, command, false)
}

func TestGetSpecCommandAndArgsNewCommandAsCommand(t *testing.T) {
	positionalArgs := []string{"sleep", "60"}
	commandArgs := []string{}
	argsArgs := []string{}
	isCommand := true
	argsLenAtDash := 0

	extraArgs, command := convertOldCommandArgsFlags(argsLenAtDash, positionalArgs, commandArgs, argsArgs, isCommand)

	assert.Equal(t, len(extraArgs), 2)
	assert.Equal(t, extraArgs[0], "sleep")
	assert.Equal(t, extraArgs[1], "60")
	assert.Equal(t, command, true)
}

func TestGetSpecCommandAndArgsBothIgnoreOld(t *testing.T) {
	positionalArgs := []string{"sleep", "60"}
	commandArgs := []string{"bash", "-c"}
	argsArgs := []string{"echo", "test"}
	isCommand := false
	argsLenAtDash := 0

	extraArgs, command := convertOldCommandArgsFlags(argsLenAtDash, positionalArgs, commandArgs, argsArgs, isCommand)

	assert.Equal(t, len(extraArgs), 2)
	assert.Equal(t, extraArgs[0], "sleep")
	assert.Equal(t, extraArgs[1], "60")
	assert.Equal(t, command, false)
}

func TestGetSpecCommandAndArgsWithMorePositionalArgument(t *testing.T) {
	positionalArgs := []string{"Test", "sleep", "60"}
	commandArgs := []string{}
	argsArgs := []string{}
	isCommand := false
	argsLenAtDash := 1

	extraArgs, command := convertOldCommandArgsFlags(argsLenAtDash, positionalArgs, commandArgs, argsArgs, isCommand)

	assert.Equal(t, len(extraArgs), 2)
	assert.Equal(t, extraArgs[0], "sleep")
	assert.Equal(t, extraArgs[1], "60")
	assert.Equal(t, command, false)
}
