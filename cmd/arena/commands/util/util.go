package util

import (
	"fmt"

	"github.com/run-ai/runai-cli/cmd/arena/types"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/config"
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

func PrintShowingJobsInNamespaceMessage(namespaceInfo types.NamespaceInfo) {
	if namespaceInfo.ProjectName != types.ALL_PROJECTS {
		if namespaceInfo.ProjectName != "" {
			fmt.Printf("Showing jobs for project %s\n", namespaceInfo.ProjectName)
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
