package template

import (
	"github.com/spf13/cobra"
)

func GenTemplateNames(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {

	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	configs, err := PrepareTemplateList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	result := make([]string, 0, len(configs))

	for _, config := range configs {
		result = append(result, config.Name)
	}

	return result, cobra.ShellCompDirectiveNoFileComp
}