package project

import (
    "context"
    "fmt"
	rsrch_api "github.com/run-ai/researcher-service/server/pkg/runai/api"
	"github.com/run-ai/runai-cli/cmd/completion"
    "github.com/run-ai/runai-cli/cmd/constants"
    "github.com/run-ai/runai-cli/pkg/authentication/assertion"
    "github.com/run-ai/runai-cli/pkg/client"
    "github.com/run-ai/runai-cli/pkg/rsrch_client"
    restclient "k8s.io/client-go/rest"
    "os"
    "sort"
    "strings"
    "text/tabwriter"
    "time"

    "github.com/run-ai/runai-cli/pkg/ui"
    commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
    "github.com/spf13/cobra"
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

func PrepareListOfProjects(restConfig *restclient.Config, includeDeleted bool) (
                    map[string]*rsrch_api.Project, int, error) {

    rs := rsrch_client.NewRsrchClient(restConfig)
    projList, err := rs.ProjectList(context.TODO(), &rsrch_client.ProjectListOptions{
        IncludeDeleted: true,
    })
    if err != nil {
        return nil, 0, err
    }

    //
    //   if --include-deleted flag is not provided, deleted projects are hidden from the output
    //   in this case we want to add a textual message notifying the user about those hidden projects
    //
    hiddenProjects := 0

    projects := make(map[string]*rsrch_api.Project)
    for idx, project := range *projList {
        if project.IsDeleted && !includeDeleted {
            hiddenProjects += 1
            continue
        }
        projects[project.Name] = &(*projList)[idx]
    }

    return projects, hiddenProjects, nil
}

func getSortedProjects(projects map[string]*rsrch_api.Project) []*rsrch_api.Project {
    projectsArray := []*rsrch_api.Project{}
    for _, project := range projects {
        projectsArray = append(projectsArray, project)
    }

    sort.Slice(projectsArray, func(i, j int) bool {
        return projectsArray[i].Name < projectsArray[j].Name
    })

    return projectsArray
}

func printProjects(infos []*rsrch_api.Project, hiddenProjects int, defaultProject string) {
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
        if  isDefault && isDeleted {
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
        Use:     "projects [--include-deleted]",
        Aliases: []string{"project"},
        Short:   "List all available projects",
        ValidArgsFunction: completion.NoArgs,
        PreRun:  commandUtil.RoleAssertion(assertion.AssertViewerRole),
        Run:     commandUtil.WrapRunCommand(runListCommand),
    }

    command.Flags().BoolVarP(&includeDeleted, "include-deleted", "d", false, "Include deleted projects")
    return command
}
