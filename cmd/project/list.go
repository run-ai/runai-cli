package project

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/run-ai/runai-cli/cmd/completion"
	"github.com/run-ai/runai-cli/cmd/constants"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	"github.com/run-ai/runai-cli/pkg/client"
	log "github.com/sirupsen/logrus"
	restclient "k8s.io/client-go/rest"

	"github.com/run-ai/runai-cli/pkg/ui"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/spf13/cobra"

	rsrch_server "github.com/run-ai/researcher-service/server/pkg/runai/api"
	rsrch_cs "github.com/run-ai/researcher-service/server/pkg/runai/client"
)

func runListCommand(cmd *cobra.Command, args []string) error {

	//
	//   obtain default project of this session
	//
	restConfig, namespace, err := client.GetRestConfig()
	if err != nil {
		return err
	}

	defaultProject := ""
	if len(namespace) > len(constants.RunaiNsProjectPrefix) {
		defaultProject = namespace[len(constants.RunaiNsProjectPrefix):]
	}

	projects, err := PrepareListOfProjects(restConfig)
	if err != nil {
		return err
	}

	//
	//   Sort the projects, so they will always appear in the same order
	//
	projectsArray := getSortedProjects(projects)

	printProjects(projectsArray, defaultProject)

	return nil
}

//
//  ask for the list of projects from the researcher service
//  parameters:
//      restConfig - pointer to kubernetes config, used for creating RS client
//  returns:
//      list of projects
//
func PrepareListOfProjects(restConfig *restclient.Config) (
	map[string]*rsrch_server.Project, error) {

	var err error
	var projList *[]rsrch_server.Project

	clientSet, ctx, err := rsrch_cs.NewCliClientFromConfig(restConfig)
	if err != nil {
		log.Errorf("Failed to create client for fetching project list: %v", err.Error())
		return nil, err
	}

	projList, err = clientSet.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	projects := make(map[string]*rsrch_server.Project)
	for idx, project := range *projList {
		projects[project.Name] = &(*projList)[idx]
	}

	return projects, nil
}

func getSortedProjects(projects map[string]*rsrch_server.Project) []*rsrch_server.Project {
	projectsArray := []*rsrch_server.Project{}
	for _, project := range projects {
		projectsArray = append(projectsArray, project)
	}

	sort.Slice(projectsArray, func(i, j int) bool {
		return projectsArray[i].Name < projectsArray[j].Name
	})

	return projectsArray
}

func printProjects(infos []*rsrch_server.Project, defaultProject string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	ui.Line(w, "PROJECT", "DEPARTMENT", "DESERVED GPUs", "INT LIMIT", "INT AFFINITY", "TRAIN AFFINITY")

	for _, info := range infos {

		deservedInfo := "-"
		if info.DeservedGpus != 0 {
			deservedInfo = fmt.Sprintf("%v", info.DeservedGpus)
		}

		interactiveJobTimeLimitFmt := "-"
		if info.InteractiveJobTimeLimitSecs != 0 {
			t := time.Duration(info.InteractiveJobTimeLimitSecs * 1000 * 1000 * 1000)
			interactiveJobTimeLimitFmt = t.String()
		}

		isDefault := info.Name == defaultProject

		name := info.Name
		if isDefault {
			name += " (default)"
		}

		departmentName := "-"
		if info.DepartmentName != "" {
			departmentName = info.DepartmentName
		}

		ui.Line(w, name, departmentName, deservedInfo, interactiveJobTimeLimitFmt,
			strings.Join(info.InteractiveNodeAffinity, ";"),
			strings.Join(info.TrainNodeAffinity, ";"))
	}

	_ = w.Flush()
}

func listCommandDEPRECATED() *cobra.Command {

	var command = &cobra.Command{
		Use:        "list",
		Short:      fmt.Sprint("List all available projects."),
		PreRun:     commandUtil.RoleAssertion(assertion.AssertViewerRole),
		Run:        commandUtil.WrapRunCommand(runListCommand),
		Deprecated: "Please use: 'runai list project' instead",
	}

	return command
}

func ListCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:               "projects [--include-deleted]",
		Aliases:           []string{"project"},
		Short:             "List all available projects",
		ValidArgsFunction: completion.NoArgs,
		PreRun:            commandUtil.RoleAssertion(assertion.AssertViewerRole),
		Run:               commandUtil.WrapRunCommand(runListCommand),
	}

	return command
}
