package submit

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

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
