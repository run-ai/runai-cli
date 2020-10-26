package submit

import (
	yaml "gopkg.in/yaml.v2"
	"os"
	"strings"
)

func applyTemplate(templateYaml string, args *submitArgs) error {
	var templateArgs submitArgs
	resolvedTemplateYaml := os.ExpandEnv(templateYaml)
	err := yaml.Unmarshal([]byte(resolvedTemplateYaml), &templateArgs)
	if err != nil {
		return err
	}

	*args = mergeTemplateSubmitArgs(*args, templateArgs)
	return nil
}

func mergeTemplateSubmitArgs(cliArgs, templateArgs submitArgs) submitArgs {
	cliArgs.EnvironmentVariable = mergeEnvironmentVariables(&cliArgs.EnvironmentVariable, &templateArgs.EnvironmentVariable)
	cliArgs.Volumes = append(cliArgs.Volumes, templateArgs.Volumes...)
	return cliArgs
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