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

	"github.com/run-ai/runai-cli/cmd/completion"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	"github.com/run-ai/runai-cli/pkg/jobs"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"

	cmdUtil "github.com/run-ai/runai-cli/cmd/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"

	// "strconv"
	"text/tabwriter"

	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/trainer"
	"github.com/run-ai/runai-cli/pkg/client"
	prom "github.com/run-ai/runai-cli/pkg/prometheus"
	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/ui"
)

// TopCommand top command
func TopCommand() *cobra.Command {
	var allNamespaces bool
	var command = &cobra.Command{
		Use:               "jobs",
		Aliases:           []string{"job"},
		Short:             "Display information about jobs in the cluster.",
		ValidArgsFunction: completion.NoArgs,
		PreRun:            commandUtil.RoleAssertion(assertion.AssertViewerRole),
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

			cmdUtil.PrintShowingJobsInNamespaceMessage(namespaceInfo, string(v1.PodRunning))

			jobs, err = trainer.GetAllJobs(kubeClient, namespaceInfo, []string{string(v1.PodRunning)})
			if err != nil {
				log.Errorf("Failed due to %v", err)
				os.Exit(1)
			}

			jobs = trainer.MakeTrainingJobOrderdByGPUCount(jobs)
			// TODO(cheyang): Support different job describer, such as MPI job/tf job describer
			topTrainingJob(kubeClient, jobs)
		},
	}

	command.Flags().BoolVarP(&allNamespaces, "all-projects", "A", false, "show all projects.")

	return command
}

var usageFormatters = map[string]ui.FormatFunction{
	"cpuusage": func(value, model interface{}) (string, error) {
		resourceUsage, ok := value.(types.ResourceUsage)
		if !ok {
			return "", fmt.Errorf("[CPUUSAGE Format]:: expecting types.ResourceUsage")
		}
		percent, err := ui.PrecantageFormat(resourceUsage.Utilization, model)
		if err != nil {
			return "", fmt.Errorf("[CPUUSAGE Format]:: failed to format utilization to percents")
		}
		if resourceUsage.Usage == 0 {
			return percent, nil
		}
		return fmt.Sprintf("%.0fm (%s)", resourceUsage.Usage*1000, percent), nil
	},
	"memoryusage": func(value, model interface{}) (string, error) {
		resourceUsage, ok := value.(types.ResourceUsage)
		if !ok {
			return "", fmt.Errorf("[MEMORYUSAGE Format]:: expecting types.ResourceUsage")
		}
		percent, err := ui.PrecantageFormat(resourceUsage.Utilization, model)
		if err != nil {
			return "", fmt.Errorf("[MEMORYUSAGE Format]:: failed to format utilization to percents")
		}
		usage, err := ui.BytesFormat(resourceUsage.Usage, model)
		if err != nil {
			return "", fmt.Errorf("[MEMORYUSAGE Format]:: failed to format usage to bytes")
		}
		return fmt.Sprintf("%s (%s)", usage, percent), nil
	},
}

func topTrainingJob(client *client.Client, jobInfoList []trainer.TrainingJob) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	promClient, err := prom.BuildPrometheusClient(client)
	if err != nil {
		log.Errorf("Error while creating prometheus client: %v", err)
	}
	rows, err := jobs.GetJobsMetrics(promClient, jobInfoList)
	if err != nil {
		log.Warnf("Error while reading jobs metrics: %v\n", err)
	}
	err = ui.CreateTable(types.JobView{}, ui.TableOpt{
		DisplayOpt: ui.DisplayOpt{
			HideAllByDefault: false,
			Hide:             []string{"Info.Status"},
		},
		Formatts: usageFormatters,
	}).Render(w, rows).Error()
	if err != nil {
		log.Errorf("Error while printing top jobs: %v", err)
	}

	_ = w.Flush()
}
