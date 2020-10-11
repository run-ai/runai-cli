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
	mpiClient "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path"

	"github.com/run-ai/runai-cli/cmd/attach"
	"github.com/run-ai/runai-cli/cmd/trainer"

	raUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/config"
	"github.com/run-ai/runai-cli/pkg/util"
	"github.com/run-ai/runai-cli/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	SubmitMpiCommand = "submit-mpi"
)

var (
	mpijob_chart string
)

func NewRunaiSubmitMPIJobCommand() *cobra.Command {
	var (
		submitArgs submitMPIJobArgs
	)

	submitArgs.Mode = "mpijob"

	var command = &cobra.Command{
		Use:     SubmitMpiCommand + " [NAME]",
		Short:   "Submit a new MPI job.",
		Aliases: []string{"mpi", "mj"},
		Args:    cobra.RangeArgs(1, 2),
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
			configValues := ""
			err = submitArgs.setCommonRun(cmd, args, kubeClient, clientset, &configValues)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			mpiClient := mpiClient.NewForConfigOrDie(kubeClient.GetRestConfig())
			err = submitMPIJob(cmd, args, &submitArgs, kubeClient, mpiClient, &configValues)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	command.Flags().IntVar(&submitArgs.NumberProcesses, "processes", 1, "Number of distributed training processes.")
	submitArgs.addCommonFlags(command)
	return command

}

type submitMPIJobArgs struct {
	// for common args
	submitArgs `yaml:",inline"`

	// for tensorboard
	NumberProcesses int `yaml:"numProcesses"` // --workers
	TotalGPUs       int `yaml:"totalGpus"`    // --workers
}

func (submitArgs *submitMPIJobArgs) prepare(args []string) (err error) {
	err = submitArgs.check()
	if err != nil {
		return err
	}
	submitArgs.TotalGPUs = submitArgs.NumberProcesses * int(*submitArgs.GPU)
	return nil
}

func (submitArgs submitMPIJobArgs) check() error {
	err := submitArgs.submitArgs.check()
	if err != nil {
		return err
	}

	if submitArgs.Image == "" {
		return fmt.Errorf("--image must be set")
	}

	return nil
}

// add k8s nodes labels
func (submitArgs *submitMPIJobArgs) addMPINodeSelectors() {
	submitArgs.addNodeSelectors()
}

// add k8s tolerations for taints
func (submitArgs *submitMPIJobArgs) addMPITolerations() {
	submitArgs.addTolerations()
}

// Submit MPIJob
func submitMPIJob(cmd *cobra.Command, args []string, submitArgs *submitMPIJobArgs, client *client.Client, mpiClient *mpiClient.Clientset,configValues *string) (err error) {
	err = submitArgs.prepare(args)
	if err != nil {
		return err
	}

	trainer := trainer.NewMPIJobTrainer(*client)
	job, err := trainer.GetTrainingJob(submitArgs.Name, submitArgs.Namespace)
	if err != nil {
		log.Debugf("Check %s exist due to error %v", submitArgs.Name, err)
	}

	if job != nil {
		return fmt.Errorf("the job %s already exists, please delete it first. use 'runai delete %s'", submitArgs.Name, submitArgs.Name)
	}

	countJobsByBaseNameFunc := func(baseName, namespace string) (int, error) {
		baseNameSelector := fmt.Sprintf("%s=%s", workflow.BaseNameLabel, baseName)
		list, err := mpiClient.KubeflowV1alpha1().MPIJobs(namespace).List(metav1.ListOptions{LabelSelector: baseNameSelector})
		if err != nil {
			return 0, err
		}
		return len(list.Items), nil
	}

	// the master is also considered as a worker
	// submitArgs.WorkerCount = submitArgs.WorkerCount - 1
	err = workflow.SubmitJob(&submitArgs.Name, submitArgs.Mode, submitArgs.Namespace, submitArgs, &submitArgs.Labels, *configValues, mpijob_chart, client.GetClientset(), countJobsByBaseNameFunc, dryRun)
	if err != nil {
		return err
	}

	fmt.Printf("The job '%s' has been submitted successfully\n", submitArgs.Name)
	fmt.Printf("You can run `%s get %s -p %s` to check the job status\n", config.CLIName, submitArgs.Name, submitArgs.Project)

	if submitArgs.Attach != nil && *submitArgs.Attach {
		if err := attach.Attach(cmd, submitArgs.Name, raUtil.IsBoolPTrue(submitArgs.StdIn), raUtil.IsBoolPTrue(submitArgs.TTY), "", attach.DefaultAttachTimeout); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	return nil
}
