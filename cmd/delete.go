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

package cmd

import (
	"fmt"
	"os"

	"github.com/run-ai/runai-cli/cmd/flags"
	cmdUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/config"
	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/util/helm"
	"github.com/run-ai/runai-cli/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewDeleteCommand
func NewDeleteCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "delete JOB_NAME",
		Short: "Delete a job and its associated pods.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			kubeClient, err := client.GetClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			namespaceInfo, err := flags.GetNamespaceToUseFromProjectFlag(cmd, kubeClient)

			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			for _, jobName := range args {
				err = deleteTrainingJob(kubeClient, jobName, namespaceInfo, "")
				if err != nil {
					log.Error(err)
				}
			}
		},
	}

	return command
}

func deleteTrainingJob(kubeClient *client.Client, jobName string, namespaceInfo types.NamespaceInfo, trainingType string) error {
	var trainingTypes []string
	// 1. Handle legacy training job
	err := helm.DeleteRelease(jobName)
	if err == nil {
		log.Infof("Delete the job %s successfully.", jobName)
		return nil
	}

	log.Debugf("%s wasn't deleted by helm due to %v", jobName, err)

	// 2. Handle training jobs created by arena
	if trainingType == "" {
		trainingTypes, err = getTrainingTypes(jobName, namespaceInfo.Namespace, kubeClient.GetClientset())

		if err != nil {
			return err
		}

		if len(trainingTypes) == 0 {
			runaiTrainer := NewRunaiTrainer(*kubeClient)
			job, err := runaiTrainer.GetTrainingJob(jobName, namespaceInfo.Namespace)
			if err == nil && !job.CreatedByCLI() {
				return fmt.Errorf("the job '%s' exists but was not created using the runai cli", jobName)
			}

			return cmdUtil.GetJobDoesNotExistsInNamespaceError(jobName, namespaceInfo)
		} else if len(trainingTypes) > 1 {
			return fmt.Errorf("There are more than 1 training jobs with the same name %s, please double check with `%s list | grep %s`. And use `%s delete %s --type` to delete the exact one.",
				jobName,
				config.CLIName,
				jobName,
				config.CLIName,
				jobName)
		}
	} else {
		trainingTypes = []string{trainingType}
	}

	err = workflow.DeleteJob(jobName, namespaceInfo.Namespace, trainingTypes[0], kubeClient.GetClientset())
	if err != nil {
		return err
	}

	fmt.Printf("The job '%s' has been deleted successfully\n", jobName)
	// (TODO: cheyang)3. Handle training jobs created by others, to implement
	return nil
}

func isKnownTrainingType(trainingType string) bool {
	for _, knownType := range knownTrainingTypes {
		if trainingType == knownType {
			return true
		}
	}

	return false
}
