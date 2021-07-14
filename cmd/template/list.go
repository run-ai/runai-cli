package template

import (
	"github.com/run-ai/runai-cli/cmd/completion"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/spf13/cobra"
)

func ListCommandDEPRECATED() *cobra.Command {
	var command = &cobra.Command{
		Use:     "templates",
		Aliases: []string{"template"},
		Short:   "List all templates.",
		PreRun:  commandUtil.RoleAssertion(assertion.AssertViewerRole),
	}

	return command
}

func ListCommandDEPRECATED2() *cobra.Command {
	var command = &cobra.Command{
		Use:               "list",
		Short:             "Display information about templates.",
		ValidArgsFunction: completion.NoArgs,
		PreRun:            commandUtil.RoleAssertion(assertion.AssertViewerRole),
	}

	return command
}
