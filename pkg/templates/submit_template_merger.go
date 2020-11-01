package templates

import (
	"reflect"
)

func MergeSubmitTemplates(base, patch SubmitTemplate) SubmitTemplate {
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
			mergedField := append(baseFieldInterface.([]string), patchValue.Field(i).Interface().([]string)...)
			baseField.Set(reflect.ValueOf(mergedField))
		default:

		}
	}
	return base
}

func mergeTemplateFields(base, patch *TemplateField) *TemplateField{
	if patch == nil {
		return base
	}
	if base == nil {
		return patch
	}
	base.Value = patch.Value
	*base.Required = *base.Required || *patch.Required
	return base
}