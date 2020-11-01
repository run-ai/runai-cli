package templates

import (
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {
	truePtr := true
	falsePtr := false
	baseSubmitTemplate := SubmitTemplate{Gpu: &TemplateField{Required: &falsePtr}, Name: &TemplateField{Value: "Tester"}, EnvVariables: []string{"a"}}
	patchSubmitTemplate := SubmitTemplate{Gpu: &TemplateField{Required: &truePtr}, Image: &TemplateField{Value: "MyTestImage"}, EnvVariables: []string{"b"}}

	result := MergeSubmitTemplates(baseSubmitTemplate, patchSubmitTemplate)

	fmt.Println(result)
}