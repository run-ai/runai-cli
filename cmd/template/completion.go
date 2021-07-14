package template

import (
	"github.com/spf13/cobra"
)

func GenTemplateNames(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {

	result := []string{"Training", "Interactive"}
	return result, cobra.ShellCompDirectiveNoFileComp
}
