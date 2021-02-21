// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package job

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"

	"github.com/run-ai/runai-cli/pkg/workflow"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/trainer"
	cmdUtil "github.com/run-ai/runai-cli/cmd/util"

	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/ui"
	"github.com/run-ai/runai-cli/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const jobInvalidStateOnCreationTimeInSeconds = 30

func ListCommand() *cobra.Command {
	var allNamespaces bool
	var command = &cobra.Command{
		Use:     "jobs",
		Aliases: []string{"job"},
		Short:   "List all jobs.",
		PreRun:  commandUtil.RoleAssertion(assertion.AssertViewerRole),
		Run: func(cmd *cobra.Command, args []string) {
			RunJobList(cmd, args, allNamespaces)
		},
	}

	command.Flags().BoolVarP(&allNamespaces, "all-projects", "A", false, "list from all projects")

	return command
}

func RunJobList(cmd *cobra.Command, args []string, allNamespaces bool) {
	kubeClient, err := client.GetClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err != nil {
		log.Errorf("Failed due to %v", err)
		os.Exit(1)
	}

	namespaceInfo, err := flags.GetNamespaceToUseFromProjectFlagIncludingAll(cmd, kubeClient, allNamespaces)

	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	cmdUtil.PrintShowingJobsInNamespaceMessage(namespaceInfo)

	jobs := []trainer.TrainingJob{}
	trainers := trainer.NewTrainers(kubeClient)
	for _, trainer := range trainers {
		if trainer.IsEnabled() {
			trainingJobs, err := trainer.ListTrainingJobs(namespaceInfo.Namespace)
			if err != nil {
				log.Errorf("Failed due to %v", err)
				os.Exit(1)
			}
			jobs = append(jobs, trainingJobs...)
		}
	}

	invalidJobs := []string{}
	jobsMap := make(map[string]bool)
	for _, job := range jobs {
		jobsMap[job.Name()] = true
	}

	configMaps, err := kubeClient.GetClientset().CoreV1().ConfigMaps(namespaceInfo.Namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("Failed due to %v", err)
		os.Exit(1)
	} else {
		for _, item := range configMaps.Items {
			if item.Labels[workflow.BaseNameLabelSelectorName] != "" {
				if jobsMap[item.Name] == false && isJobCreationTimePass(&item) {
					invalidJobs = append(invalidJobs, item.Name)
				}
			}
		}
	}

	jobs = trainer.MakeTrainingJobOrderdByAge(jobs)

	displayTrainingJobList(jobs, invalidJobs)

}

func isJobCreationTimePass(configMap *v1.ConfigMap) bool {
	return time.Now().Sub(configMap.CreationTimestamp.Time).Seconds() > jobInvalidStateOnCreationTimeInSeconds
}

func displayTrainingJobList(jobInfoList []trainer.TrainingJob, invalidJobs []string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	labelField := []string{"NAME", "STATUS", "AGE", "NODE", "IMAGE", "TYPE", "PROJECT", "USER", "GPUs Allocated (Requested)", "PODs Running (Pending)", "SERVICE URL(S)"}

	ui.Line(w, labelField...)

	for _, jobInfo := range jobInfoList {

		status := GetJobRealStatus(jobInfo)
		nodeName := jobInfo.HostIPOfChief()
		if strings.Contains(nodeName, ", ") {
			nodeName = "<multiple>"
		}

		// For backward compatability. Indicat jobs on default namespace
		var projectName string
		if jobInfo.Namespace() == "default" {
			projectName = fmt.Sprintf("%s (old)", jobInfo.Project())
		} else {
			projectName = jobInfo.Project()
		}

		currentAllocatedGPUs := jobInfo.CurrentAllocatedGPUs()
		currentAllocatedGPUsAsString := fmt.Sprintf("%g", currentAllocatedGPUs)
		if currentAllocatedGPUs == 0 && trainer.IsFinishedStatus(status) {
			currentAllocatedGPUsAsString = "-"
		}
		allocatedFromRequestedGPUs := fmt.Sprintf("%s (%v)", currentAllocatedGPUsAsString, jobInfo.RequestedGPUString())
		runningOfActivePods := fmt.Sprintf("%d (%d)", int(jobInfo.RunningPods()), int(jobInfo.PendingPods()))

		ui.Line(w, jobInfo.Name(),
			status,
			util.ShortHumanDuration(jobInfo.Age()),
			nodeName, jobInfo.Image(), jobInfo.Trainer(), projectName, jobInfo.User(),
			allocatedFromRequestedGPUs,
			runningOfActivePods,
			strings.Join(jobInfo.ServiceURLs(), ", "))
	}

	for _, invalidJob := range invalidJobs {
		ui.Line(w, invalidJob, "Invalid job", "", "", "", "", "", "", "", "", "")
	}
	_ = w.Flush()
}
