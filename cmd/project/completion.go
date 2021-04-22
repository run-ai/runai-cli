package project

import (
	"github.com/run-ai/runai-cli/cmd/completion"
	"github.com/spf13/cobra"
)

const COMPLETION_PROJ_FILE_SUFFIX = "proj"

func GenProjectNamesForFlag(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {

	result := completion.ReadFromCache(COMPLETION_PROJ_FILE_SUFFIX)
	if result != nil {
		return result, cobra.ShellCompDirectiveNoFileComp
	}

	projects, err := PrepareListOfProjects();
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	projectNames := make([]string, 0, len(projects))
	for projectName, _ := range projects {
		projectNames = append(projectNames, projectName)
	}

	completion.WriteToCache(COMPLETION_PROJ_FILE_SUFFIX, projectNames)

	return projectNames, cobra.ShellCompDirectiveNoFileComp
}

func GenProjectNamesForArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

	//
	//    for arg - we have to prevent duplicate values
	//
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return GenProjectNamesForFlag(cmd, args, toComplete)
}

