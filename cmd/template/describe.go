package template

import (
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/spf13/cobra"
)

func getCommandDEPRECATED() *cobra.Command {
	var command = &cobra.Command{
		Use:    "get TEMPLATE_NAME",
		Short:  "Get information about one of the templates in the cluster.",
		PreRun: commandUtil.RoleAssertion(assertion.AssertViewerRole),
	}

	return command
}

func DescribeCommandDEPRECATED() *cobra.Command {
	var command = &cobra.Command{
		Use:               "template [TEMPLATE_NAME]",
		Args:              cobra.RangeArgs(1, 1),
		Short:             "Describe information about one of the templates in the cluster.",
		ValidArgsFunction: GenTemplateNames,
		PreRun:            commandUtil.RoleAssertion(assertion.AssertViewerRole),
	}

	return command
}
