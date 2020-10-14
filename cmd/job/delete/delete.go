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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"strings"

	"github.com/run-ai/runai-cli/cmd/trainer"

	"github.com/run-ai/runai-cli/cmd/flags"
	cmdUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/config"
	"github.com/run-ai/runai-cli/pkg/types"
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
				maybeJobIdentifier := JobIdentifier{name: jobName, namespace: namespaceInfo.Namespace}
				err = DeleteJob(maybeJobIdentifier, kubeClient)
				if err != nil {
					log.Error(err)
				}
				err = deleteTrainingJob(kubeClient, jobName, namespaceInfo)
				if err != nil {
					log.Error(err)
				}
			}
		},
	}

	return command
}

func getJobOptionalConfigMaps(name, namespace string, clientset kubernetes.Interface) ([]string, error) {
	var configMaps []string
	configMapInNamespace, err := clientset.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, trainingType := range trainer.KnownTrainingTypes {
		configMapPrefix := fmt.Sprintf("%s-%s", name, trainingType)
		for _, configMap := range configMapInNamespace.Items {
			if strings.HasPrefix(configMap.Name, configMapPrefix){
				configMaps = append(configMaps, configMap.Name)
			}
		}
	}
	return configMaps, nil
}

func deleteTrainingJob(kubeClient *client.Client, jobName string, namespaceInfo types.NamespaceInfo) error {
	optionalConfigMaps, err := getJobOptionalConfigMaps(jobName, namespaceInfo.Namespace, kubeClient.GetClientset())
	if err != nil {
		return err
	}
	if len(optionalConfigMaps) == 0 {
		runaiTrainer := trainer.NewRunaiTrainer(*kubeClient)
		job, err := runaiTrainer.GetTrainingJob(jobName, namespaceInfo.Namespace)
		if err == nil && !job.CreatedByCLI() {
			return fmt.Errorf("the job '%s' exists but was not created using the runai cli", jobName)
		}
		return cmdUtil.GetJobDoesNotExistsInNamespaceError(jobName, namespaceInfo)
	} else if len(optionalConfigMaps) > 1 {
		return fmt.Errorf("There are more than 1 training jobs with the same name %s, please double check with `%s list | grep %s`. And use `%s delete %s --type` to delete the exact one.",
			jobName,
			config.CLIName,
			jobName,
			config.CLIName,
			jobName)
	}

	err = workflow.DeleteJob(namespaceInfo.Namespace, optionalConfigMaps[0], kubeClient.GetClientset())
	if err != nil {
		return err
	}

	fmt.Printf("The job '%s' has been deleted successfully\n", jobName)
	// (TODO: cheyang)3. Handle training jobs created by others, to implement
	return nil
}


