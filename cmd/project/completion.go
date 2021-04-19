package project

import "github.com/spf13/cobra"

func GenProjectNamesForFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projects, err := PrepareListOfProjects();
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	projectNames := make([]string, 0, len(projects))
	for projectName, _ := range projects {
		projectNames = append(projectNames, projectName)
	}

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

