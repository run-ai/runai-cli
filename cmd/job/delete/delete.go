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
	"context"
	"fmt"
	rsrch_server "github.com/run-ai/researcher-service/server/pkg/runai/api"
	rsrch_cs "github.com/run-ai/researcher-service/server/pkg/runai/client"
	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/job"
	"github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/rsrch_client"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

// NewDeleteCommand
func NewDeleteCommand() *cobra.Command {
	var isAll bool

	var command = &cobra.Command{
		Use:               "delete JOB_NAME",
		Short:             "Delete a job and its associated pods.",
		ValidArgsFunction: job.GenJobNames,
		PreRun:            commandUtil.NamespacedRoleAssertion(assertion.AssertExecutorRole),
		Run: func(cmd *cobra.Command, args []string) {
			if !isAll && len(args) == 0 {
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

			projectName := util.ToProject(namespaceInfo.Namespace)

			//
			//    obtain the list of job names to delete, based on the command's arguments
			//
			jobNamesToDelete := args

			if isAll {
				jobNamesToDelete, err = job.ListJobNamesByNamespace(kubeClient, namespaceInfo)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			}

			restConfig, _, err := client.GetRestConfig()
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}

			//
			//   prepare the request as a list of job names + project
			//
			jobsToDelete := make([]rsrch_server.DeletedJob, 0, len(args))

			for _, jobNameToDelete := range jobNamesToDelete {
				jobsToDelete = append(jobsToDelete, rsrch_server.DeletedJob{
					Name:    jobNameToDelete,
					Project: projectName,
				})
			}

			//
			//    connect to the researcher config, if it can serve delete job request
			//
			var deleteJobsStatus []rsrch_server.DeletedJobStatus

			rs := rsrch_client.NewRsrchClient(restConfig, rsrch_client.DeleteJobMinVersion)
			if rs != nil {
				//
				//   RS can serve the request, so send it to RS
				//
				deleteJobsStatus, err = rs.JobDelete(context.TODO(), jobsToDelete)
			} else {
				log.Infof("researcher-service cannot serve the request, use in-house CLI code for job delete")

				clientSet, err := rsrch_cs.NewCliClientFromConfig(restConfig)
				if err != nil {
					log.Errorf("Failed to create clientSet for in-house CLI job delete: %v", err.Error())
					return
				}

				deleteJobsStatus = clientSet.DeleteJobs(context.TODO(), jobsToDelete)
			}

			if err != nil {
				log.Error(err)
				fmt.Printf("Error occured while attempting to delete jobs.\n")
			} else {
				for _, deleteJobStatus := range deleteJobsStatus {
					if deleteJobStatus.Ok {
						fmt.Printf("Job %s deleted successfully.\n", deleteJobStatus.Name)
					} else if deleteJobStatus.Error.Status == http.StatusNotFound {
						fmt.Printf("Job %s does not exist in project %s. If the job exists in a different project, use -p <project-name>.\n", deleteJobStatus.Name, projectName)
					} else {
						log.Errorf("%v: %v", deleteJobStatus.Error.Message, deleteJobStatus.Error.Details)
						fmt.Printf("Failed to delete job %s: %s\n", deleteJobStatus.Name, deleteJobStatus.Error.Message)
					}
				}
			}

		},
	}

	command.Flags().BoolVarP(&isAll, "all", "A", false, "Delete all jobs")

	return command
}
