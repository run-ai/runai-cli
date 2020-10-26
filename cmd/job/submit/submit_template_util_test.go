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