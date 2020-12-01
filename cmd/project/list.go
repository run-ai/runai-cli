package project

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/auth"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/ui"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	projectResource = schema.GroupVersionResource{
		Group:    "run.ai",
		Version:  "v1",
		Resource: "projects",
	}
)

type ProjectInfo struct {
	name                        string
	deservedGPUs                string
	defaultProject              bool
	interactiveJobTimeLimitSecs string
	nodeAffinityInteractive     string
	nodeAffinityTraining        string
	department                  string
}

type Project struct {
	Spec struct {
		DeservedGpus                 float64  `mapstructure:"deservedGpus,omitempty"`
		InteractiveJobTimeLimitSecs  int      `mapstructure:"interactiveJobTimeLimitSecs,omitempty"`
		Department                   string   `json:"department,omitempty" protobuf:"bytes,1,opt,name=department"`
		NodeAffinityInteractive     []string  `json:"nodeAffinityInteractive,omitempty" protobuf:"bytes,1,opt,name=nodeAffinityInteractive"`
		NodeAffinityTrain           []string  `json:"nodeAffinityTrain,omitempty" protobuf:"bytes,1,opt,name=nodeAffinityTrain"`
	} `mapstructure:"spec,omitempty"`
	Metadata struct {
		Name string `mapstructure:"name,omitempty"`
	} `mapstructure:"metadata,omitempty"`
}

func runListCommand(cmd *cobra.Command, args []string) error {
	kubeClient, err := client.GetClient()
	if err != nil {
		return err
	}

	projects := make(map[string]*ProjectInfo)

	dynamicClient, err := dynamic.NewForConfig(kubeClient.GetRestConfig())
	if err != nil {
		return err
	}

	projectList, err := dynamicClient.Resource(projectResource).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, projectItem := range projectList.Items {
		var project Project

		if err := mapstructure.Decode(projectItem.Object, &project); err != nil {
			return err
		}

		projects[project.Metadata.Name] = &ProjectInfo{
			name:                        project.Metadata.Name,
			defaultProject:              kubeClient.GetDefaultNamespace() == "runai-"+project.Metadata.Name,
			deservedGPUs:                fmt.Sprintf("%.2f", project.Spec.DeservedGpus),
			interactiveJobTimeLimitSecs: strconv.Itoa(project.Spec.InteractiveJobTimeLimitSecs),
			nodeAffinityInteractive:     strings.Join(project.Spec.NodeAffinityInteractive, ","),
			nodeAffinityTraining:        strings.Join(project.Spec.NodeAffinityTrain, ","),
			department:                  project.Spec.Department,
		}
	}

	// Sort the projects, so they will always appear in the same order
	projectsArray := getSortedProjects(projects)

	printProjects(projectsArray)

	return nil
}

func getSortedProjects(projects map[string]*ProjectInfo) []*ProjectInfo {
	projectsArray := []*ProjectInfo{}
	for _, project := range projects {
		projectsArray = append(projectsArray, project)
	}

	sort.Slice(projectsArray, func(i, j int) bool {
		return projectsArray[i].name < projectsArray[j].name
	})

	return projectsArray
}

func printProjects(infos []*ProjectInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	ui.Line(w, "PROJECT", "DEPARTMENT", "DESERVED GPUs", "INT LIMIT", "INT AFFINITY", "TRAIN AFFINITY")

	for _, info := range infos {
		deservedInfo := "deleted"

		if info.deservedGPUs != "" {
			deservedInfo = info.deservedGPUs
		}

		interactiveJobTimeLimitFmt := "-"
		if info.interactiveJobTimeLimitSecs != "" && info.interactiveJobTimeLimitSecs != "0" {
			i, _ := strconv.Atoi(info.interactiveJobTimeLimitSecs)
			t := time.Duration(i * 1000 * 1000 * 1000)
			interactiveJobTimeLimitFmt = t.String()
		}

		var name string
		if info.defaultProject {
			name = fmt.Sprintf("%s (default)", info.name)
		} else {
			name = info.name
		}

		ui.Line(w, name, info.department, deservedInfo, interactiveJobTimeLimitFmt, info.nodeAffinityInteractive, info.nodeAffinityTraining)
	}

	_ = w.Flush()
}

func listCommandDEPRECATED() *cobra.Command {

	var command = &cobra.Command{
		Use:        "list",
		Short:      fmt.Sprint("List all available projects."),
		PreRun: 	commandUtil.RoleAssertion(auth.AssertViewerRole),
		Run:        commandUtil.WrapRunCommand(runListCommand),
		Deprecated: "Please use: 'runai list project' instead",
	}

	return command
}

func ListCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:     "projects",
		Aliases: []string{"project"},
		Short:   "List all available projects",
		PreRun:  commandUtil.RoleAssertion(auth.AssertViewerRole),
		Run:     commandUtil.WrapRunCommand(runListCommand),
	}

	return command
}
