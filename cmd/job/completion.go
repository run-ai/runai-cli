package job

import (
	"github.com/run-ai/runai-cli/cmd/completion"
	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/project"
	"github.com/run-ai/runai-cli/cmd/template"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	"os"
)

const CompletionJobsFileSuffix = "jobs"
const CompletionPodsFileSuffix = "pods_"

//
//   generate job names for commands which require job name as parameter
//
func GenJobNames(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {

	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	kubeClient, err := client.GetClient()
	if err != nil {
		log.Errorf("Failed due to %v", err)
		os.Exit(1)
	}

	namespaceInfo, err := flags.GetNamespaceInfoToUse(cmd, kubeClient)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	cachePath := CompletionJobsFileSuffix + "." + namespaceInfo.ProjectName

	result := completion.ReadFromCache(cachePath)
	if result != nil {
		return result, cobra.ShellCompDirectiveNoFileComp
	}

	jobs, invalidJobs, err := PrepareTrainerJobList(kubeClient, namespaceInfo)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	result = make([]string, 0, len(jobs))

	for _, curJob := range(jobs) {
		result = append(result, curJob.Name())
	}

	for _, invalidJob := range(invalidJobs) {
		result = append(result, invalidJob)
	}

	completion.WriteToCache(cachePath, result)

	return result, cobra.ShellCompDirectiveNoFileComp
}

//
//   generate completion list of pod names for a given job.
//   Assumption: in all the commands that has --pod parameter, the first argument is the job name
//
func GenPodNames(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {

	if len(args) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	//
	//    pods are cahced on a per job basis cause user can change job name while typing thr command
	//    and in this case we need to re-load the informaiton of the new job
	//
	cachePath := CompletionPodsFileSuffix + args[0]
	result := completion.ReadFromCache(cachePath)
	if result != nil {
		return result, cobra.ShellCompDirectiveNoFileComp
	}

	jobInfo, _, err := PrepareJobInfo(cmd, args[0])
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	result = make([]string, 0, len(jobInfo.AllPods()))
	for _, curPod := range(jobInfo.AllPods()) {
		result = append(result, curPod.Name)
	}

	completion.WriteToCache(cachePath, result)

	return result, cobra.ShellCompDirectiveNoFileComp
}

//
//   add pod flag to the command, and register compleiton function for it
//
func AddPodNameFlag(cmd *cobra.Command, retValue *string) {
	cmd.Flags().StringVar(retValue, "pod", "", "Specify a pod of a running job. To get a list of the pods of a specific job, run \"runai describe <job-name>\" command")
	cmd.RegisterFlagCompletionFunc("pod", GenPodNames)
}

//
//   add description for submit and submit-mpi flags .Note that some flags are relevant only for one of those two
//   commands, but we don't care as the flag registration function ignores flags which are not supported by the command
//
func AddSubmitFlagsCompletion(command *cobra.Command) {
	completion.AddFlagDescrpition(command, "backoff-limit", "Specify the number of times the job will be retried before failing")
	completion.AddFlagDescrpition(command, "completions", "Specify the number of successful pods required for this job to be completed")
	completion.AddFlagDescrpition(command, "cpu", "Specify number of CPU units to allocate (e.g. 0.5, 1)")
	completion.AddFlagDescrpition(command, "cpu-limit", "Specify CPU limit for the job (e.g. 0.5, 1)")
	completion.AddFlagDescrpition(command, "create-home-dir", "Specify a temporary home directory to be created")
	completion.AddFlagDescrpition(command, "environment", "Specify values for environment variable, formatted as 'variable=value'")
	completion.AddFlagDescrpition(command, "git-sync", "Specify sync string var1=value1;var2=value2;...")
	completion.AddFlagDescrpition(command, "gpu", "Specify GPU units to allocate (e.g. 0.5, 1)")
	completion.AddFlagDescrpition(command, "gpu-memory", "Specify GPU memory to allocate (e.g. 1G, 500M)")
	completion.AddFlagDescrpition(command, "image", "Specify image to use when creating the job")
	completion.AddFlagDescrpition(command, "job-name-prefix", "Specify prefix for the job name")
	completion.AddFlagDescrpition(command, "memory", "Specify CPU memory to allocate (e.g. 1G, 20M)")
	completion.AddFlagDescrpition(command, "memory-limit", "Specify memory limit (e.g. 1G, 20M)")
	completion.AddFlagDescrpition(command, "name", "Specify a name for the job")
	completion.AddFlagDescrpition(command, "node-type", "Specify node-type label for enforcing node type affinity")
	completion.AddFlagDescrpition(command, "parallelism", "Specify number of pods to run in parallel at any given time")
	completion.AddFlagDescrpition(command, "port", "Specify ports to expose from the job container")
	completion.AddFlagDescrpition(command, "processes","Specify number of distributed training processes")
	completion.AddFlagDescrpition(command, "pvc", "Specify mount parameters of a persistent volume")
	completion.AddFlagDescrpition(command, "ttl-after-finish", "Specify the auto-deletion duration (e.g. 2s, 5m, 3h)")
	completion.AddFlagDescrpition(command, "volume", "Specify volumes to mount, formatted as '<host_path>:<container_path>:<access_mode>'")
	completion.AddFlagDescrpition(command, "working-dir", "Specify the working directory of the container")

	command.RegisterFlagCompletionFunc("template", template.GenTemplateNames)
	command.RegisterFlagCompletionFunc("image-pull-policy", completion.ImagePolicyValues)
	command.RegisterFlagCompletionFunc("service-type", completion.ServiceTypeValues)

	command.RegisterFlagCompletionFunc(flags.ProjectFlag, project.GenProjectNamesForFlag)
}