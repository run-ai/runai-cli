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
	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/jobs"
	"github.com/run-ai/runai-cli/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// NewDeleteCommand
func NewDeleteCommand() *cobra.Command {
	var interactive string
	var trainerType string
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
				maybeJobIdentifier := jobs.JobIdentifier{Name: jobName, Namespace: namespaceInfo.Namespace, Trainer: strings.ToLower(trainerType), Interactive: strings.ToLower(interactive)}
				err = DeleteJob(maybeJobIdentifier, kubeClient)
				if err != nil {
					log.Error(err)
				}
			}
		},
	}

	command.Flags().StringVarP(&interactive, "interactive", "", "unknown", "Specifies whether to delete interactive job [interactive / train]")
	command.Flags().StringVarP(&trainerType, "trainer-type", "", "", "Specifies the trainer type to avoid conflict")
	return command
}

func DeleteJob(maybeJobIdentifier jobs.JobIdentifier, kubeClient *client.Client) error {
	trainingJob, err := jobs.GetTrainingJob(maybeJobIdentifier, kubeClient)
	if err != nil {
		return err
	}

	if !trainingJob.CreatedByCLI() {
		return fmt.Errorf("the job '%s' exists but was not created using the runai cli", trainingJob.Name())
	}

	err = workflow.DeleteJob(trainingJob, kubeClient)
	if err != nil {
		return err
	}

	fmt.Printf("The job '%s' has been deleted successfully\n", trainingJob.Name())
	return nil
}

