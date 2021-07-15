package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	rsrch_server "github.com/run-ai/researcher-service/server/pkg/runai/api"
	rsrch_cs "github.com/run-ai/researcher-service/server/pkg/runai/client"
	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/job"
	"github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	"github.com/run-ai/runai-cli/pkg/client"
	pkgUtil "github.com/run-ai/runai-cli/pkg/util"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type directCommand func(rsrch_server.Interface, context.Context, []rsrch_server.ResourceID) []rsrch_server.JobActionStatus

// NewSuspendCommand creates a new suspend command for cobra to suspend jobs.
func NewSuspendCommand() *cobra.Command {
	var isAll bool

	var command = &cobra.Command{
		Use:               "suspend JOB_NAME",
		Short:             "Suspend a job and its associated pods.",
		ValidArgsFunction: job.GenJobNames,
		PreRun:            commandUtil.NamespacedRoleAssertion(assertion.AssertExecutorRole),
		Run: func(cmd *cobra.Command, args []string) {
			suspendWorkflowHelper(cmd, args, rsrch_server.Interface.SuspendJobs, "suspend", isAll)
		},
	}

	command.Flags().BoolVarP(&isAll, "all", "A", false, "Suspend all jobs")

	return command
}

// NewResumeCommand creates a new resume command for cobra to resume jobs.
func NewResumeCommand() *cobra.Command {
	var isAll bool

	var command = &cobra.Command{
		Use:               "resume JOB_NAME",
		Short:             "Resume a job and its associated pods.",
		ValidArgsFunction: job.GenJobNames,
		PreRun:            commandUtil.NamespacedRoleAssertion(assertion.AssertExecutorRole),
		Run: func(cmd *cobra.Command, args []string) {
			suspendWorkflowHelper(cmd, args, rsrch_server.Interface.ResumeJobs, "resume", isAll)
		},
	}

	command.Flags().BoolVarP(&isAll, "all", "A", false, "Resume all jobs")

	return command
}

func suspendWorkflowHelper(cmd *cobra.Command, args []string, directCmd directCommand, cmdName string, isAll bool) {
	if !isAll && len(args) == 0 {
		cmd.HelpFunc()(cmd, args)
		os.Exit(1)
	}

	kubeClient, err := client.GetClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if !pkgUtil.CheckComponentVersion("runai-job-controller", ">=v0.1.8", kubeClient) {
		fmt.Printf("runai job controller version should be >=0.1.8\n")
		os.Exit(1)
	}

	namespaceInfo, err := flags.GetNamespaceToUseFromProjectFlag(cmd, kubeClient)
	if err != nil {
		log.Debugf("Failed due to %v", err)
		fmt.Println(err)
		os.Exit(1)
	}

	projectName := util.ToProject(namespaceInfo.Namespace)
	jobNamesToSuspend := args

	if isAll {
		jobNamesToSuspend, err = job.ListJobNamesByNamespace(kubeClient, namespaceInfo)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}
	jobs := make([]rsrch_server.ResourceID, 0, len(jobNamesToSuspend))
	for _, jobName := range jobNamesToSuspend {
		jobs = append(jobs, rsrch_server.ResourceID{
			Name:    jobName,
			Project: projectName,
		})
	}

	clientSet, ctx, err := rsrch_cs.NewCliClientFromConfig(kubeClient.GetRestConfig())
	if err != nil {
		log.Errorf("Failed to create clientSet for modifying CLI job %s: %v", cmdName, err.Error())
		return
	}
	cmdStatuses := directCmd(clientSet, ctx, jobs)
	for _, status := range cmdStatuses {
		if status.Ok {
			// trim trailing 'e' from cmd name to convert to past tense
			if cmdName[len(cmdName)-1] == 'e' {
				cmdName = cmdName[:len(cmdName)-1]
			}
			fmt.Printf("Job %s %sed successfully.\n", status.Name, cmdName)
		} else if status.Error != nil {
			if status.Error.Status == http.StatusNotFound {
				fmt.Printf("Job %s not found \n", status.Name)
			} else {
				fmt.Printf("Job %s failed to %s: %s\n", status.Name, cmdName, status.Error.Message)
			}
		}
	}
}
