package templates

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

var (
	truePtr = true
	falsePtr = false
)

func TestMergeTemplateFieldsOverrideRequired(t *testing.T) {
	base := TemplateField{Required: &falsePtr}
	patch := TemplateField{Required: &truePtr}

	merged := mergeTemplateFields(&base, &patch)

	assert.Equal(t, *merged.Required, true)
}

func TestMergeTemplateFieldsOverrideValue(t *testing.T) {
	base := TemplateField{Value: "4"}
	patch := TemplateField{Value: "5"}

	merged := mergeTemplateFields(&base, &patch)

	assert.Equal(t, merged.Value, "5")
}

func TestMergeTemplateFieldsOverrideNil(t *testing.T) {
	base := TemplateField{}
	patch := TemplateField{Required: &truePtr}

	merged := mergeTemplateFields(&base, &patch)

	assert.Equal(t, *merged.Required, true)
}

func TestMergeTemplateFieldsUnOverrideNil(t *testing.T) {
	base := TemplateField{Required: &truePtr}
	patch := TemplateField{}

	merged := mergeTemplateFields(&base, &patch)

	assert.Equal(t, *merged.Required, true)
}


func TestMergeSubmitTemplatesSanity(t *testing.T) {
	baseSubmitTemplate := SubmitTemplate{Gpu: &TemplateField{Required: &falsePtr}, Name: &TemplateField{Value: "Tester"}, EnvVariables: []string{"a"}}
	patchSubmitTemplate := SubmitTemplate{Gpu: &TemplateField{Required: &truePtr}, Image: &TemplateField{Value: "MyTestImage"}, EnvVariables: []string{"b"}}

	result := mergeSubmitTemplates(baseSubmitTemplate, patchSubmitTemplate)

	assert.Equal(t, *result.Gpu.Required, true)
	assert.Equal(t, result.Name.Value, "Tester")
	assert.Equal(t, result.Image.Value, "MyTestImage")
	assert.Equal(t, len(result.EnvVariables), 2)
}