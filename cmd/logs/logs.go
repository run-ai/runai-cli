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

package logs

import (
	"fmt"
	"github.com/run-ai/runai-cli/cmd/trainer"
	"os"
	"path"
	"time"

	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/pkg/client"
	podlogs "github.com/run-ai/runai-cli/pkg/podlogs"
	tlogs "github.com/run-ai/runai-cli/pkg/printer/base/logs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewLogsCommand() *cobra.Command {
	var outerArgs = &podlogs.OuterRequestArgs{}
	var command = &cobra.Command{
		Use:   "logs JOB_NAME",
		Short: "Print the logs of a job.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			name := args[0]

			kubeClient, err := client.GetClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			clientset := kubeClient.GetClientset()
			namespaceInfo, err := flags.GetNamespaceToUseFromProjectFlag(cmd, kubeClient)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			outerArgs.KubeClient = clientset
			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			// podName, err := getPodNameFromJob(printer.kubeClient, namespace, name)
			job, err := trainer.SearchTrainingJob(kubeClient, name, "", namespaceInfo)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			outerArgs.Namespace = namespaceInfo.Namespace
			outerArgs.RetryCount = 5
			outerArgs.RetryTimeout = time.Millisecond
			names := []string{}
			for _, pod := range job.AllPods() {
				names = append(names, path.Base(pod.ObjectMeta.SelfLink))
			}
			chiefPod := job.ChiefPod()
			if len(names) > 1 && outerArgs.PodName == "" {
				names = []string{path.Base(chiefPod.ObjectMeta.SelfLink)}
			}
			logPrinter, err := tlogs.NewPodLogPrinter(names, outerArgs)
			if err != nil {
				log.Errorf(err.Error())
				os.Exit(1)
			}
			code, err := logPrinter.Print()
			if err != nil {
				log.Errorf("%s, %s", err.Error(), "please use \"runai get\" to get more information.")
				os.Exit(1)
			} else if code != 0 {
				os.Exit(code)
			}
		},
	}

	command.Flags().BoolVarP(&outerArgs.Follow, "follow", "f", false, "Stream the logs.")
	command.Flags().DurationVar(&outerArgs.SinceSeconds, "since", 0, "Return logs newer than a relative duration, like 5s, 2m, or 3h. Note that only one flag \"since-time\" or \"since\" may be used.")
	command.Flags().StringVar(&outerArgs.SinceTime, "since-time", "", "Return logs after a specific date (e.g. 2019-10-12T07:20:50.52Z). Note that only one flag \"since-time\" or \"since\" may be used.")
	command.Flags().IntVarP(&outerArgs.Tail, "tail", "t", -1, "Return a specific number of log lines.")
	command.Flags().BoolVar(&outerArgs.Timestamps, "timestamps", false, "Include timestamps on each line in the log output.")

	// command.Flags().StringVar(&printer.pod, "instance", "", "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().StringVar(&outerArgs.PodName, "pod", "", "Specify a pod of a running job. To get a list of the pods of a specific job, run \"runai describe <job-name>\" command")
	return command
}
