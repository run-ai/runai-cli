package submit

import (
	"github.com/run-ai/runai-cli/pkg/templates"
	"strings"
)

func applyTemplate(templateYaml string, args *submitArgs) error {
	template, err := templates.GetSubmitTemplateFromYaml(templateYaml, true)
	if err != nil {
		return err
	}

	*args = mergeTemplateToSubmitArgs(*args, template)
	return nil
}

func mergeTemplateToSubmitArgs(args submitArgs, template *templates.SubmitTemplate) submitArgs {
	args.EnvironmentVariable = mergeEnvironmentVariables(&args.EnvironmentVariable, &template.EnvVariables)
	args.Volumes = append(args.Volumes, template.Volumes...)
	return args
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