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

package submit

import (
	"fmt"
	"github.com/run-ai/runai-cli/cmd/completion"
	"github.com/run-ai/runai-cli/cmd/job"
	"os"
	"path"
	"strconv"

	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"

	"github.com/run-ai/runai-cli/cmd/attach"
	"github.com/run-ai/runai-cli/cmd/flags"
	raUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/config"
	"github.com/run-ai/runai-cli/pkg/util"
	"github.com/run-ai/runai-cli/pkg/workflow"
	"github.com/spf13/cobra"
)

const (
	SubmitMpiCommand = "submit-mpi"
	mpiExamples      = `
runai submit-mpi --name distributed-job --processes=2 -g 1 \
	-i gcr.io/run-ai-demo/quickstart-distributed`
)

var (
	mpijob_chart string
)

func NewRunaiSubmitMPIJobCommand() *cobra.Command {
	var (
		submitArgs submitMPIJobArgs
	)

	var command = &cobra.Command{
		Use:     SubmitMpiCommand + " [NAME]",
		Short:   "Submit a new MPI job.",
		Aliases: []string{"mpi", "mj"},
		Example: mpiExamples,
		ValidArgsFunction: completion.NoArgs,
		PreRun:  commandUtil.NamespacedRoleAssertion(assertion.AssertExecutorRole),
		Run: func(cmd *cobra.Command, args []string) {
			kubeClient, err := client.GetClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			chartPath, err := util.GetChartsFolder()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			mpijob_chart = path.Join(chartPath, "mpijob")

			clientset := kubeClient.GetClientset()

			commandArgs := convertOldCommandArgsFlags(cmd, &submitArgs.submitArgs, args)
			submitArgs.GitSync = GitSyncFromConnectionString(gitSyncConnectionString)

			err = applyTemplate(&submitArgs, commandArgs, clientset)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = submitArgs.setCommonRun(cmd, args, kubeClient, clientset)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if len(submitArgs.Image) == 0 {
				fmt.Print("\n-i, --image must be set\n\n")
				os.Exit(1)
			}

			err = submitMPIJob(cmd, args, &submitArgs, kubeClient)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	fbg := flags.NewFlagsByGroups(command)
	submitArgs.addCommonFlags(fbg)
	fg := fbg.GetOrAddFlagSet(JobLifecycleFlagGroup)
	flags.AddIntNullableFlag(fg, &(submitArgs.Processes), "processes", "Number of distributed training processes.")

	fbg.UpdateFlagsByGroupsToCmd()

	job.AddSubmitFlagsCompletion(command)

	return command

}

type submitMPIJobArgs struct {
	// for common args
	submitArgs `yaml:",inline"`

	// for tensorboard
	Processes       *int    // --workers
	NumberProcesses int     `yaml:"numProcesses"` // --workers
	TotalGPUs       float64 `yaml:"totalGpus"`    // --workers
	TotalGPUsMemory int     `yaml:"totalGpusMemory"`
}

func (submitArgs *submitMPIJobArgs) prepare(args []string) (err error) {
	err = submitArgs.check()
	if err != nil {
		return err
	}
	numberProcesses := 1
	if submitArgs.Processes != nil {
		numberProcesses = *submitArgs.Processes
	}

	gpus := float64(0)
	if submitArgs.GPU != nil {
		gpus = *submitArgs.GPU
	}
	submitArgs.TotalGPUs = float64(numberProcesses) * gpus

	gpusMemory := uint64(0)
	if parsedGpusMemory, err := strconv.ParseUint(submitArgs.GPUMemory, 10, 64); err == nil {
		gpusMemory = parsedGpusMemory
	}
	submitArgs.TotalGPUsMemory = numberProcesses * int(gpusMemory)

	submitArgs.NumberProcesses = numberProcesses
	return nil
}

func (submitArgs submitMPIJobArgs) check() error {
	err := submitArgs.submitArgs.check()
	if err != nil {
		return err
	}

	return nil
}

// Submit MPIJob
func submitMPIJob(cmd *cobra.Command, args []string, submitArgs *submitMPIJobArgs, client *client.Client) (err error) {
	err = submitArgs.prepare(args)
	if err != nil {
		return err
	}

	// the master is also considered as a worker
	// submitArgs.WorkerCount = submitArgs.WorkerCount - 1
	submitArgs.Name, err = workflow.SubmitJob(submitArgs.Name, submitArgs.Namespace, submitArgs.generateSuffix, submitArgs, mpijob_chart, client.GetClientset(), dryRun)
	if err != nil {
		return err
	}

	if !dryRun {
		fmt.Printf("The job '%s' has been submitted successfully\n", submitArgs.Name)
		fmt.Printf("You can run `%s describe job %s -p %s` to check the job status\n", config.CLIName, submitArgs.Name, submitArgs.Project)

		if submitArgs.Attach != nil && *submitArgs.Attach {
			if err := attach.Attach(cmd, submitArgs.Name, raUtil.IsBoolPTrue(submitArgs.StdIn), raUtil.IsBoolPTrue(submitArgs.TTY), "", attach.DefaultAttachTimeout); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	return nil
}
