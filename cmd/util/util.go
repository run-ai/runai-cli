package util

import (
	"fmt"

	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/config"
	"github.com/run-ai/runai-cli/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNamespaceFromProjectName(project string, kubeClient *client.Client) (string, error) {
	namespaceList, err := kubeClient.GetClientset().CoreV1().Namespaces().List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", RUNAI_QUEUE_LABEL, project),
	})

	if err != nil {
		return "", err
	}

	if namespaceList != nil && len(namespaceList.Items) != 0 {
		return namespaceList.Items[0].Name, nil
	} else {
		return "", fmt.Errorf("project %s was not found. Please run '%s project list' to view all avaliable projects", project, config.CLIName)
	}
}

func GetJobDoesNotExistsInNamespaceError(jobName string, namespaceInfo types.NamespaceInfo) error {
	if namespaceInfo.ProjectName != "" {
		return fmt.Errorf("The job %s does not exist in project %s. If the job exists in a different project, use -p <project name>.", jobName, namespaceInfo.ProjectName)
	} else {
		return fmt.Errorf("The job %s does not exist in backward compatability mode. If the job exists in a specific project, use -p <project name>.", jobName)
	}
}

func PrintShowingJobsInNamespaceMessageByStatuses(namespaceInfo types.NamespaceInfo, status v1.PodPhase) {
	if namespaceInfo.ProjectName != types.AllProjects {
		if namespaceInfo.ProjectName != "" {
			var statusMessage = ""
			if len(status) != 0 {
				statusMessage = fmt.Sprintf("with status %s ", status)
			}
			fmt.Printf("Showing jobs %sfor project %s\n", statusMessage, namespaceInfo.ProjectName)
		} else {
			fmt.Println("Showing old jobs")
		}
	}
}

func BoolP(b bool) *bool {
	return &b
}

func IsBoolPTrue(b *bool) bool {
	return b != nil && *b
}
