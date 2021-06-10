package cmd

import (
	"fmt"
	"os"

	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/job"
	runaiClient "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/types"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"github.com/run-ai/runai-cli/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type workflowCommand func(string, types.NamespaceInfo, runaiClient.Interface) error

// NewSuspendCommand creates a new suspend command for cobra to suspend jobs.
func NewSuspendCommand() *cobra.Command {
	var isAll bool

	var command = &cobra.Command{
		Use:               "suspend JOB_NAME",
		Short:             "Suspend a job and its associated pods.",
		ValidArgsFunction: job.GenJobNames,
		PreRun:            commandUtil.NamespacedRoleAssertion(assertion.AssertExecutorRole),
		Run: func(cmd *cobra.Command, args []string) {
			suspendWorkflowHelper(cmd, args, workflow.SuspendJob, isAll)
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
			suspendWorkflowHelper(cmd, args, workflow.ResumeJob, isAll)
		},
	}

	command.Flags().BoolVarP(&isAll, "all", "A", false, "Resume all jobs")

	return command
}

func suspendWorkflowHelper(cmd *cobra.Command, args []string, workflowCmd workflowCommand, isAll bool) {
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

	jobNamesToSuspend := args

	if isAll {
		jobNamesToSuspend, err = job.ListJobNamesByNamespace(kubeClient, namespaceInfo)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}

	runaijobClient := runaiClient.NewForConfigOrDie(kubeClient.GetRestConfig())
	for _, jobName := range jobNamesToSuspend {
		err = workflowCmd(jobName, namespaceInfo, runaijobClient)
		if err != nil {
			log.Error(err)
		}
	}
}
