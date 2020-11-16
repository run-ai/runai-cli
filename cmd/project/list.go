package project

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/mitchellh/mapstructure"
	constants "github.com/run-ai/runai-cli/cmd/constants"
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

func runListCommand(cmd *cobra.Command, args []string) error {
	kubeClient, err := client.GetClient()
	if err != nil {
		return err
	}

	clientset := kubeClient.GetClientset()

	namespaceList, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})

	if err != nil {
		return err
	}

	projects := make(map[string]*ProjectInfo)

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

	dynamicClient, err := dynamic.NewForConfig(kubeClient.GetRestConfig())
	if err != nil {
		return err
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

func newListProjectsCommand_DEPRECATED() *cobra.Command {

	var command = &cobra.Command{
		Use:   "list",
		Short: "List all avaliable projects. DEPRECATED! use instead > runai list project.",
		Run:  commandUtil.WrapRunCommand(runListCommand),
	}

	return command
}

func NewListProjectCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:   "project",
		Aliases: []string{"projects"},
		Short: "List all avaliable projects",
		Run:   commandUtil.WrapRunCommand(runListCommand),
	}

	return command
}
