package project

import (
	"fmt"
	"github.com/run-ai/runai-cli/cmd/completion"
	"github.com/run-ai/runai-cli/cmd/constants"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
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
	queueResource = schema.GroupVersionResource{
		Group:    "scheduling.incubator.k8s.io",
		Version:  "v1alpha1",
		Resource: "queues",
	}

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

// Required for backwards compatibility with clusters that don't have the Project resource yet.
type Queue struct {
	Spec struct {
		DeservedGpus                 float64  `mapstructure:"deservedGpus,omitempty"`
		InteractiveJobTimeLimitSecs  int      `mapstructure:"interactiveJobTimeLimitSecs,omitempty"`
		Department                   string   `json:"department,omitempty" protobuf:"bytes,1,opt,name=department"`
		NodeAffinityInteractiveTypes []string `json:"nodeAffinityInteractiveTypes,omitempty" protobuf:"bytes,1,opt,name=nodeAffinityInteractiveTypes"`
		NodeAffinityTrainTypes       []string `json:"nodeAffinityTrainTypes,omitempty" protobuf:"bytes,1,opt,name=nodeAffinityTrainTypes"`
	} `mapstructure:"spec,omitempty"`
	Metadata struct {
		Name string `mapstructure:"name,omitempty"`
	} `mapstructure:"metadata,omitempty"`
}

type Project struct {
	Spec struct {
		DeservedGpus                float64  `mapstructure:"deservedGpus,omitempty"`
		InteractiveJobTimeLimitSecs int      `mapstructure:"interactiveJobTimeLimitSecs,omitempty"`
		Department                  string   `json:"department,omitempty" protobuf:"bytes,1,opt,name=department"`
		NodeAffinityInteractive     []string `json:"nodeAffinityInteractive,omitempty" protobuf:"bytes,1,opt,name=nodeAffinityInteractive"`
		NodeAffinityTrain           []string `json:"nodeAffinityTrain,omitempty" protobuf:"bytes,1,opt,name=nodeAffinityTrain"`
	} `mapstructure:"spec,omitempty"`
	Metadata struct {
		Name string `mapstructure:"name,omitempty"`
	} `mapstructure:"metadata,omitempty"`
}

func runListCommand(cmd *cobra.Command, args []string) error {

	projects, err := PrepareListOfProjects();
	if err != nil {
		return err
	}

	// Sort the projects, so they will always appear in the same order
	projectsArray := getSortedProjects(projects)
	printProjects(projectsArray)
	return nil
}

func PrepareListOfProjects() (map[string]*ProjectInfo, error) {
	kubeClient, err := client.GetClient()
	if err != nil {
		return nil, err
	}

	projects := make(map[string]*ProjectInfo)

	dynamicClient, err := dynamic.NewForConfig(kubeClient.GetRestConfig())
	if err != nil {
		return nil, err
	}

	if err := listProjects(dynamicClient, projects, kubeClient.GetDefaultNamespace()); err != nil {
		return nil, err
	} else if len(projects) < 1 {
		// If list projects didn't populate anything in the map fallback to listing queues
		if err = listQueues(dynamicClient, kubeClient, projects); err != nil {
			return nil ,err
		}
	}

	return projects, nil
}

func listProjects(dynamicClient dynamic.Interface, projects map[string]*ProjectInfo, defaultNamespace string) error {
	projectList, err := dynamicClient.Resource(projectResource).List(metav1.ListOptions{})
	if errors.IsNotFound(err) {
		log.Debug(err)
		// Cluster doesn't know about the 'Project' resource - fallback to listing queues.
		return nil
	}

	for _, projectItem := range projectList.Items {
		var project Project

		if err := mapstructure.Decode(projectItem.Object, &project); err != nil {
			return err
		}

		projects[project.Metadata.Name] = &ProjectInfo{
			name:                        project.Metadata.Name,
			defaultProject:              defaultNamespace == "runai-"+project.Metadata.Name,
			deservedGPUs:                fmt.Sprintf("%.2f", project.Spec.DeservedGpus),
			interactiveJobTimeLimitSecs: strconv.Itoa(project.Spec.InteractiveJobTimeLimitSecs),
			nodeAffinityInteractive:     strings.Join(project.Spec.NodeAffinityInteractive, ","),
			nodeAffinityTraining:        strings.Join(project.Spec.NodeAffinityTrain, ","),
			department:                  project.Spec.Department,
		}
	}
	return nil
}

// listQueues Provides backwards compatibility for clusters that didn't upgrade to using Projects.
// This is the exact same code that was here before I changed the list command logic.
func listQueues(dynamicClient dynamic.Interface, kubeClient *client.Client, projects map[string]*ProjectInfo) error {
	clientset := kubeClient.GetClientset()
	namespaceList, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, namespace := range namespaceList.Items {
		if namespace.Labels == nil {
			continue
		}
		runaiQueue := namespace.Labels[constants.RUNAI_QUEUE_LABEL]
		if runaiQueue != "" {
			projects[runaiQueue] = &ProjectInfo{
				name:           runaiQueue,
				defaultProject: kubeClient.GetDefaultNamespace() == namespace.Name,
			}
		}
	}

	queueList, err := dynamicClient.Resource(queueResource).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, queueItem := range queueList.Items {
		var queue Queue
		if err := mapstructure.Decode(queueItem.Object, &queue); err != nil {
			return err
		}
		if project, found := projects[queue.Metadata.Name]; found {
			project.deservedGPUs = fmt.Sprintf("%.2f", queue.Spec.DeservedGpus)
			project.interactiveJobTimeLimitSecs = strconv.Itoa(queue.Spec.InteractiveJobTimeLimitSecs)
			project.nodeAffinityInteractive = strings.Join(queue.Spec.NodeAffinityInteractiveTypes, ",")
			project.nodeAffinityTraining = strings.Join(queue.Spec.NodeAffinityTrainTypes, ",")
			project.department = queue.Spec.Department
		}
	}
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
		PreRun:     commandUtil.RoleAssertion(assertion.AssertViewerRole),
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
		ValidArgsFunction: completion.NoArgs,
		PreRun:  commandUtil.RoleAssertion(assertion.AssertViewerRole),
		Run:     commandUtil.WrapRunCommand(runListCommand),
	}

	return command
}
