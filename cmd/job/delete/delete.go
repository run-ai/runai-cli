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
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	"github.com/run-ai/runai-cli/pkg/client"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/run-ai/runai-cli/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

// NewDeleteCommand
func NewDeleteCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:    "delete JOB_NAME",
		Short:  "Delete a job and its associated pods.",
		PreRun: commandUtil.NamespacedRoleAssertion(assertion.AssertExecutorRole),
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
				err = workflow.DeleteJob(jobName, namespaceInfo, kubeClient.GetClientset())
				if err != nil {
					log.Error(err)
				}
			}
		},
	}

	return command
}
