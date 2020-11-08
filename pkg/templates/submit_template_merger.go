package templates

import (
	"reflect"
	"strings"
)

const environmentVariableTemplateFieldName = "EnvVariables"

func MergeSubmitTemplatesYamls(baseYaml, patchYaml string) (*SubmitTemplate, error) {
	baseTemplate, err := GetSubmitTemplateFromYaml(baseYaml)
	if err != nil {
		return nil, err
	}

	patchTemplate, err := GetSubmitTemplateFromYaml(patchYaml)
	if err != nil {
		return nil, err
	}

	mergedTemplate := mergeSubmitTemplates(*baseTemplate, *patchTemplate)
	return &mergedTemplate, nil
}

func mergeSubmitTemplates(base, patch SubmitTemplate) SubmitTemplate {
	basePtrValue := reflect.ValueOf(&base)
	baseValue := basePtrValue.Elem()

	patchValue := reflect.ValueOf(patch)

	for i := 0; i < baseValue.NumField(); i++ {
		baseField := baseValue.Field(i)

		baseFieldInterface := baseField.Interface()
		switch baseFieldInterface.(type) {
		case *TemplateField:
			mergedField := mergeTemplateFields(baseFieldInterface.(*TemplateField), patchValue.Field(i).Interface().(*TemplateField))
			baseField.Set(reflect.ValueOf(mergedField))
		case []string:
			var mergedField []string
			if baseValue.Type().Field(i).Name == environmentVariableTemplateFieldName {
				baseArray := baseFieldInterface.([]string)
				patchArray := patchValue.Field(i).Interface().([]string)
				mergedField = MergeEnvironmentVariables(&patchArray, &baseArray)
			} else {
				mergedField = append(baseFieldInterface.([]string), patchValue.Field(i).Interface().([]string)...)
			}
			baseField.Set(reflect.ValueOf(mergedField))
		default:

		}
	}
	return base
}

func mergeTemplateFields(base, patch *TemplateField) *TemplateField {
	if patch == nil {
		return base
	}
	if base == nil {
		return patch
	}
	base.Value = patch.Value

	if patch.Required != nil {
		base.Required = patch.Required
	}
	return base
}

func MergeEnvironmentVariables(baseEnvVars, patchEnvVar *[]string) []string {
	cliEnvVarMap := make(map[string]bool)

	for _, cliVar := range *baseEnvVars {
		maybeKeyVal := strings.Split(cliVar, "=")
		if len(maybeKeyVal) != 2 {
			continue
		}
		key := maybeKeyVal[0]
		cliEnvVarMap[key] = true
	}

	for _, templateVar := range *patchEnvVar {
		maybeKeyVal := strings.Split(templateVar, "=")
		if len(maybeKeyVal) != 2 {
			continue
		}
		key := maybeKeyVal[0]
		if !cliEnvVarMap[key] {
			*baseEnvVars = append(*baseEnvVars, templateVar)
		}
	}

	return *baseEnvVars
}
