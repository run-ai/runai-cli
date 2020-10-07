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

package top

import (
	"fmt"
	"io"
	"os"
	"strings"

	cmdUtil "github.com/run-ai/runai-cli/cmd/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"

	"strconv"
	"text/tabwriter"

	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/trainer"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/util"
)

func NewTopJobCommand() *cobra.Command {
	var allNamespaces bool
	var command = &cobra.Command{
		Use:   "job",
		Short: "Display information about jobs in the cluster.",
		Run: func(cmd *cobra.Command, args []string) {

			kubeClient, err := client.GetClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			
			namespaceInfo, err := flags.GetNamespaceToUseFromProjectFlagIncludingAll(cmd, kubeClient, allNamespaces)

			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			var (
				jobs []trainer.TrainingJob
			)

			cmdUtil.PrintShowingJobsInNamespaceMessage(namespaceInfo)

			trainers := trainer.NewTrainers(kubeClient)
			for _, trainer := range trainers {
				trainingJobs, err := trainer.ListTrainingJobs(namespaceInfo.Namespace)
				if err != nil {
					log.Errorf("Failed due to %v", err)
					os.Exit(1)
				}

				for _, job := range trainingJobs {
					if job.GetStatus() != string(v1.PodSucceeded) {
						jobs = append(jobs, job)
					}
				}
			}

			jobs = trainer.MakeTrainingJobOrderdByGPUCount(jobs)
			// TODO(cheyang): Support different job describer, such as MPI job/tf job describer
			topTrainingJob(jobs)
		},
	}

	command.Flags().BoolVarP(&allNamespaces, "all-projects", "A", false, "show all projects.")

	return command
}

func topTrainingJob(jobInfoList []trainer.TrainingJob) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		totalAllocatedGPUs float64
		totalRequestedGPUs float64
	)

	labelField := []string{"NAME", "PROJECT", "GPU(Current Requests)", "GPU(Current Allocated)", "STATUS", "TYPE", "AGE", "NODE"}

	PrintLine(w, labelField...)

	for _, jobInfo := range jobInfoList {

		hostIP := jobInfo.HostIPOfChief()
		requestedGPU := jobInfo.CurrentRequestedGPUs()
		allocatedGPU := jobInfo.CurrentAllocatedGPUs()
		// status, hostIP := jobInfo.getStatus()
		totalAllocatedGPUs += allocatedGPU
		totalRequestedGPUs += requestedGPU
		PrintLine(w, jobInfo.Name(),
			jobInfo.Project(),
			strconv.FormatFloat(jobInfo.CurrentRequestedGPUs(), 'f', -1, 64),
			strconv.FormatFloat(jobInfo.CurrentAllocatedGPUs(), 'f', -1, 64),
			jobInfo.GetStatus(),
			jobInfo.Trainer(),
			util.ShortHumanDuration(jobInfo.Age()),
			hostIP,
		)
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Total Allocated GPUs: ")
	fmt.Fprintf(w, "%v \t\n", strconv.FormatFloat(totalAllocatedGPUs, 'f', -1, 32))
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Total Requested GPUs: ")
	fmt.Fprintf(w, "%s \t\n", strconv.FormatFloat(totalRequestedGPUs, 'f', -1, 32))

	_ = w.Flush()
}

func fromByteToMiB(value float64) float64 {
	return value / 1048576
}

// todo remove to ui
func PrintLine(w io.Writer, fields ...string) {
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}
