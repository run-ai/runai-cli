package project

import (
	"context"
	"fmt"
	"github.com/run-ai/runai-cli/cmd/completion"
	"github.com/run-ai/runai-cli/cmd/constants"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/rsrch_client"
	log "github.com/sirupsen/logrus"
	restclient "k8s.io/client-go/rest"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/run-ai/runai-cli/pkg/ui"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/spf13/cobra"

	rsrch_server "github.com/run-ai/researcher-service/server/pkg/runai/api"
	rsrch_cs "github.com/run-ai/researcher-service/server/pkg/runai/client"
)

var includeDeleted bool

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

	projects, hiddenProjects, err := PrepareListOfProjects(restConfig, includeDeleted)
	if err != nil {
		return err
	}

	//
	//   Sort the projects, so they will always appear in the same order
	//
	projectsArray := getSortedProjects(projects)

	printProjects(projectsArray, hiddenProjects, defaultProject)

	return nil
}

//
//  ask for the list of projects from the researcher service
//  parameters:
//      restConfig - pointer to kubernetes config, used for creating RS client
//      includeDeleted - indication if we want the result list to include deleted projects
//  returns:
//      list of projects
//      number of hidden proejects (=deleted projects which are filtered out from the list)
//
func PrepareListOfProjects(restConfig *restclient.Config, includeDeleted bool) (
	map[string]*rsrch_server.Project, int, error) {

	rs := rsrch_client.NewRsrchClient(restConfig, rsrch_client.ProjectListMinVersion)

	var err error
	var projList *[]rsrch_server.Project

	if rs != nil {
		projList, err = rs.ProjectList(context.TODO(), &rsrch_client.ProjectListOptions{
			// even if deleted projects are filtered out, we still count the number
			// of deleted projects, thus needs the entire list
			IncludeDeleted: true,
		})
	} else {
		log.Infof("RS cannot serve the request, use in-house CLI code for project list")

		clientSet, err := rsrch_cs.NewCliClientFromConfig(restConfig)
		if err != nil {
			log.Errorf("Failed to create clientSet for in-house CLI project list: %v", err.Error())
			return nil, 0, err
		}

		projList, err = clientSet.GetProjects(context.TODO(), true)
	}

	if err != nil {
		return nil, 0, err
	}

	//
	//   if --include-deleted flag is not provided, deleted projects are hidden from the output
	//   in this case we want to add a textual message notifying the user about those hidden projects
	//
	hiddenProjects := 0

	projects := make(map[string]*rsrch_server.Project)
	for idx, project := range *projList {
		if project.IsDeleted && !includeDeleted {
			hiddenProjects += 1 // don't include in the list, only count them
		} else {
			projects[project.Name] = &(*projList)[idx] // include in the list
		}
	}

	return projects, hiddenProjects, nil
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

func printProjects(infos []*rsrch_server.Project, hiddenProjects int, defaultProject string) {
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
		isDeleted := info.IsDeleted

		name := info.Name
		if isDefault && isDeleted {
			name += " (default,deleted)"
		} else if isDefault {
			name += " (default)"
		} else if isDeleted {
			name += " (deleted)"
		}

		departmentName := "-"
		if info.DepartmentName != "" {
			departmentName = info.DepartmentName
		}

		ui.Line(w, name, departmentName, deservedInfo, interactiveJobTimeLimitFmt,
			strings.Join(info.InteractiveNodeAffinity, ";"),
			strings.Join(info.TrainNodeAffinity, ";"))
	}

	if hiddenProjects != 0 {
		hiddenMsg := ""
		if hiddenProjects == 1 {
			hiddenMsg = fmt.Sprintf("\nAdditionally, there is 1 deleted project. Use the --include-deleted flag to show it.\n")
		} else {
			hiddenMsg = fmt.Sprintf("\nAdditionally, there are %d deleted projects. Use the --include-deleted flag to show them.\n", hiddenProjects)
		}
		w.Write([]byte(hiddenMsg))
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

	command.Flags().BoolVarP(&includeDeleted, "include-deleted", "d", false, "Include deleted projects")
	return command
}
