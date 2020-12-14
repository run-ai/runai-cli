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
	"github.com/run-ai/runai-cli/pkg/auth"
	"os/user"
	"strconv"
	"syscall"

	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/global"
	raUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/clusterConfig"
	"github.com/run-ai/runai-cli/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
)

const (
	runaiNamespace = "runai"
	jobDefaultName = "job"
	dashArg        = "--"
	commandFlag    = "command"
	oldCommandFlag = "old-command"

	// flag group names
	AliasesAndShortcutsFlagGroup flags.FlagGroupName = "Aliases/Shortcuts"
	ContainerDefinitionFlagGroup flags.FlagGroupName = "Container Definition"
	ResourceAllocationFlagGroup  flags.FlagGroupName = "Resource Allocation"
	StorageFlagGroup             flags.FlagGroupName = "Storage"
	NetworkFlagGroup             flags.FlagGroupName = "Network"
	JobLifecycleFlagGroup        flags.FlagGroupName = "Job Lifecycle"
	AccessControlFlagGroup       flags.FlagGroupName = "Access Control"
	SchedulingFlagGroup          flags.FlagGroupName = "Scheduling"
)

var (
	dryRun                  bool
	templateName            string
	gitSyncConnectionString string
)

// The common parts of the submitAthd
type submitArgs struct {
	Image               string `yaml:"image"`
	NameParameter       string
	Project             string `yaml:"project,omitempty"`
	Interactive         *bool  `yaml:"interactive,omitempty"`
	User                string `yaml:"user,omitempty"`
	Name                string
	Namespace           string
	GPU                 *float64 `yaml:"gpu,omitempty"`
	GPUInt              *int     `yaml:"gpuInt,omitempty"`
	GPUFraction         string   `yaml:"gpuFraction,omitempty"`
	NodeType            string   `yaml:"node_type,omitempty"`
	SpecArgs            []string `yaml:"args,omitempty"`
	CPU                 string   `yaml:"cpu,omitempty"`
	CPULimit            string   `yaml:"cpuLimit,omitempty"`
	Memory              string   `yaml:"memory,omitempty"`
	MemoryLimit         string   `yaml:"memoryLimit,omitempty"`
	EnvironmentVariable []string `yaml:"environment,omitempty"`

	ImagePullPolicy            string   `yaml:"imagePullPolicy"`
	AlwaysPullImage            *bool    `yaml:"alwaysPullImage,omitempty"`
	Volumes                    []string `yaml:"volume,omitempty"`
	PersistentVolumes          []string `yaml:"persistentVolumes,omitempty"`
	WorkingDir                 string   `yaml:"workingDir,omitempty"`
	PreventPrivilegeEscalation *bool    `yaml:"preventPrivilegeEscalation"`
	CreateHomeDir              *bool    `yaml:"createHomeDir,omitempty"`
	RunAsUser                  string   `yaml:"runAsUser,omitempty"`
	RunAsGroup                 string   `yaml:"runAsGroup,omitempty"`
	SupplementalGroups         []int    `yaml:"supplementalGroups,omitempty"`
	RunAsCurrentUser           *bool
	SpecCommand                []string          `yaml:"command"`
	Command                    *bool             `yaml:"isCommand"`
	LocalImage                 *bool             `yaml:"localImage,omitempty"`
	LargeShm                   *bool             `yaml:"shm,omitempty"`
	Ports                      []string          `yaml:"ports,omitempty"`
	Labels                     map[string]string `yaml:"labels,omitempty"`
	HostIPC                    *bool             `yaml:"hostIPC,omitempty"`
	HostNetwork                *bool             `yaml:"hostNetwork,omitempty"`
	StdIn                      *bool             `yaml:"stdin,omitempty"`
	TTY                        *bool             `yaml:"tty,omitempty"`
	Attach                     *bool             `yaml:"attach,omitempty"`
	NamePrefix                 string            `yaml:"namePrefix,omitempty"`
	BackoffLimit               *int              `yaml:"backoffLimit,omitempty"`
	GitSync                    *GitSync          `yaml:"gitSync,omitempty"`
	generateSuffix             bool
}

func (s submitArgs) check() error {
	if s.Name == "" {
		return fmt.Errorf("--name must be set")
	}

	// return fmt.Errorf("must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.")
	err := util.ValidateJobName(s.Name)
	if err != nil {
		return err
	}

	return nil
}

func (submitArgs *submitArgs) addCommonFlags(fbg flags.FlagsByGroups) {

	flagSet := fbg.GetOrAddFlagSet(AliasesAndShortcutsFlagGroup)
	flagSet.StringVar(&submitArgs.NameParameter, "name", "", "Job name")
	flags.AddBoolNullableFlag(flagSet, &(submitArgs.Interactive), "interactive", "", "Mark this Job as interactive.")
	flagSet.StringVarP(&(templateName), "template", "", "", "Use a specific template to run this job (otherwise use the default template if exists).")
	flagSet.StringVarP(&(submitArgs.Project), "project", "p", "", "Specifies a project. Set a default project using 'runai config project <project name>'.")
	// Will not submit the job to the cluster, just print the template to the screen
	flagSet.BoolVar(&dryRun, "dry-run", false, "Run as dry run")
	flagSet.MarkHidden("dry-run")
	flagSet.StringVar(&submitArgs.NamePrefix, "job-name-prefix", "", "Set defined prefix for the job name and add index as suffix")

	flagSet = fbg.GetOrAddFlagSet(ContainerDefinitionFlagGroup)
	flagSet.StringVar(&(submitArgs.ImagePullPolicy), "image-pull-policy", "Always", "set image pull policy: always, ifNotPresent or never.")
	flags.AddBoolNullableFlag(flagSet, &(submitArgs.AlwaysPullImage), "always-pull-image", "", "Always pull latest version of the image.")
	flagSet.MarkDeprecated("always-pull-image", "please use 'image-pull-policy=Always' instead.")
	flagSet.StringArrayVar(&(submitArgs.SpecArgs), "args", []string{}, "Arguments to pass to the command run on container start. Use together with --command.")
	flagSet.MarkDeprecated("args", "please use -- with extra arguments. See usage")
	flagSet.StringArrayVarP(&(submitArgs.EnvironmentVariable), "environment", "e", []string{}, "Set environment variables in the container.")
	flagSet.StringVarP(&(submitArgs.Image), "image", "i", "", "Container image to use when creating the job.")
	flagSet.StringArrayVar(&(submitArgs.SpecCommand), oldCommandFlag, []string{}, "Run this command on container start. Use together with --args.")
	flagSet.MarkHidden(oldCommandFlag)
	flags.AddBoolNullableFlag(flagSet, &submitArgs.Command, commandFlag, "", "If true, overrides the image's entrypoint with the command supplied after '--'.")
	flags.AddBoolNullableFlag(flagSet, &submitArgs.LocalImage, "local-image", "", "Use an image stored locally on the machine running the job.")
	flagSet.MarkDeprecated("local-image", "please use 'image-pull-policy=Never' instead.")
	flags.AddBoolNullableFlag(flagSet, &submitArgs.TTY, "tty", "t", "Allocate a TTY for the container.")
	flags.AddBoolNullableFlag(flagSet, &submitArgs.StdIn, "stdin", "", "Keep stdin open on the container(s) in the pod, even if nothing is attached.")
	flags.AddBoolNullableFlag(flagSet, &submitArgs.Attach, "attach", "", `If true, wait for the Pod to start running, and then attach to the Pod as if 'runai attach ...' were called. Attach makes tty and stdin true by default. Default false`)
	flagSet.StringVar(&(submitArgs.WorkingDir), "working-dir", "", "Set the container's working directory.")
	flagSet.StringVar(&gitSyncConnectionString, "git-sync", "", "sync string as explained in the documentation")
	flags.AddBoolNullableFlag(flagSet, &(submitArgs.RunAsCurrentUser), "run-as-user", "", "Run in the context of the current CLI user rather than the root user.")

	flagSet = fbg.GetOrAddFlagSet(ResourceAllocationFlagGroup)
	flags.AddFloat64NullableFlagP(flagSet, &(submitArgs.GPU), "gpu", "g", "GPU units to allocate for the Job (0.5, 1).")
	flagSet.StringVar(&(submitArgs.CPU), "cpu", "", "CPU units to allocate for the job (0.5, 1)")
	flagSet.StringVar(&(submitArgs.Memory), "memory", "", "CPU Memory to allocate for this job (1G, 20M)")
	flagSet.StringVar(&(submitArgs.CPULimit), "cpu-limit", "", "CPU limit for the job (0.5, 1)")
	flagSet.StringVar(&(submitArgs.MemoryLimit), "memory-limit", "", "Memory limit for this job (1G, 20M)")
	flags.AddBoolNullableFlag(flagSet, &submitArgs.LargeShm, "large-shm", "", "Mount a large /dev/shm device.")

	flagSet = fbg.GetOrAddFlagSet(StorageFlagGroup)
	flagSet.StringArrayVarP(&(submitArgs.Volumes), "volume", "v", []string{}, "Volumes to mount into the container.")
	flagSet.StringArrayVar(&(submitArgs.PersistentVolumes), "pvc", []string{}, "Mount a persistent volume. Syntax: 'StorageClass[optional]:Size:ContainerMountPath:ro[optional]")
	flagSet.StringArrayVar(&(submitArgs.Volumes), "volumes", []string{}, "Volumes to mount into the container.")
	flagSet.MarkDeprecated("volumes", "please use 'volume' flag instead.")

	flagSet = fbg.GetOrAddFlagSet(NetworkFlagGroup)
	flags.AddBoolNullableFlag(flagSet, &(submitArgs.HostIPC), "host-ipc", "", "Use the host's ipc namespace.")
	flags.AddBoolNullableFlag(flagSet, &(submitArgs.HostNetwork), "host-network", "", "Use the host's network stack inside the container.")

	flagSet = fbg.GetOrAddFlagSet(JobLifecycleFlagGroup)
	flags.AddIntNullableFlag(flagSet, &(submitArgs.BackoffLimit), "backoff-limit", "The number of times the job will be retried before failing. Default 6.")
	flags.AddIntNullableFlag(flagSet, &(submitArgs.BackoffLimit), "backoffLimit", "The number of times the job will be retried before failing. Default 6.")
	flagSet.MarkDeprecated("backoffLimit", "use backoff-limit instead")

	flagSet = fbg.GetOrAddFlagSet(AccessControlFlagGroup)
	flags.AddBoolNullableFlag(flagSet, &submitArgs.CreateHomeDir, "create-home-dir", "", "Create a temporary home directory. Default is true when the --run-as-user flag is set, and false if not.")
	flags.AddBoolNullableFlag(flagSet, &(submitArgs.PreventPrivilegeEscalation), "prevent-privilege-escalation", "", "Prevent the job’s container from gaining additional privileges after start.")
	flagSet.StringVarP(&(submitArgs.User), "user", "u", "", "Use different user to run the Job.")
	flagSet.MarkHidden("user")

	flagSet = fbg.GetOrAddFlagSet(SchedulingFlagGroup)
	flagSet.StringVar(&(submitArgs.NodeType), "node-type", "", "Enforce node type affinity by setting a node-type label.")
}

func (submitArgs *submitArgs) setCommonRun(cmd *cobra.Command, args []string, kubeClient *client.Client, clientset kubernetes.Interface) error {
	util.SetLogLevel(global.LogLevel)
	assignUser(submitArgs)
	name, generateSuffix, err := getJobNameWithSuffixGenerationFlag(cmd, args, submitArgs)
	if err != nil {
		return err
	}
	submitArgs.generateSuffix = generateSuffix

	var errs = validation.IsDNS1035Label(name)
	if len(errs) > 0 {
		fmt.Println("")
		return fmt.Errorf("Job names must consist of lower case alphanumeric characters or '-' and start with an alphabetic character (e.g. 'my-name',  or 'abc-123')")
	}

	submitArgs.Name = name

	namespaceInfo, err := flags.GetNamespaceToUseFromProjectFlagAndPrintError(cmd, kubeClient)

	if err != nil {
		return err
	}

	if namespaceInfo.ProjectName == "" {
		return fmt.Errorf("Define a project by --project flag, alternatively set a project as default")
	}

	clusterConfig, err := clusterConfig.GetClusterConfig(clientset)
	if err != nil {
		return err
	}

	submitArgs.Namespace = namespaceInfo.Namespace
	submitArgs.Project = namespaceInfo.ProjectName
	if clusterConfig.EnforceRunAsUser || raUtil.IsBoolPTrue(submitArgs.RunAsCurrentUser) {
		currentUser, err := user.Current()
		if err != nil {
			return fmt.Errorf("Could not retrieve the current user: %s", err.Error())
		}

		groups, err := syscall.Getgroups()
		if err != nil {
			return fmt.Errorf("Could not retrieve list of groups for user: %s", err.Error())

		}
		submitArgs.SupplementalGroups = groups
		submitArgs.RunAsUser = currentUser.Uid
		submitArgs.RunAsGroup = currentUser.Gid
		// Set the default of CreateHomeDir as true if run-as-user is true
		// todo: not set as true until testing it
		// if submitArgs.CreateHomeDir == nil {
		// 	t := true
		// 	submitArgs.CreateHomeDir = &t
		// }
	}

	if clusterConfig.EnforcePreventPrivilegeEscalation {
		preventPrivilegeEscalation := true
		submitArgs.PreventPrivilegeEscalation = &preventPrivilegeEscalation
	}

	err = HandleVolumesAndPvc(submitArgs)
	if err != nil {
		return err
	}

	index, err := getJobIndex(clientset)

	if err != nil {
		log.Debug("Could not get job index. Will not set a label.")
	} else {
		submitArgs.Labels = make(map[string]string)
		submitArgs.Labels["runai/job-index"] = index
	}

	// by default when the user set --attach the --stdin and --tty set to true
	if raUtil.IsBoolPTrue(submitArgs.Attach) {
		if submitArgs.StdIn == nil {
			submitArgs.StdIn = raUtil.BoolP(true)
		}

		if submitArgs.TTY == nil {
			submitArgs.TTY = raUtil.BoolP(true)
		}
	}

	if raUtil.IsBoolPTrue(submitArgs.TTY) && !raUtil.IsBoolPTrue(submitArgs.StdIn) {
		return fmt.Errorf("--stdin is required for containers with -t/--tty=true")
	}

	handleRequestedGPUs(submitArgs)
	if err = handleImagePullPolicy(submitArgs); err != nil {
		return err
	}

	if err = submitArgs.GitSync.HandleGitSync(); err != nil {
		return err
	}
	return nil
}

func assignUser(submitArgs *submitArgs) {
	// If its not zero then it was passed as flag by the user
	if submitArgs.User == "" {
		if kubeLoginUser, err := auth.GetEmailForCurrentKubeloginToken(); err == nil && kubeLoginUser != "" {
			// Try to get user from kubelogin cached token
			submitArgs.User = kubeLoginUser
		} else if osUser, err := user.Current(); err == nil {
			// Fallback to OS user
			submitArgs.User = osUser.Username
		}
	}
}

func getJobIndex(clientset kubernetes.Interface) (string, error) {
	for true {
		index, shouldTryAgain, err := tryGetJobIndexOnce(clientset)

		if index != "" || !shouldTryAgain {
			return index, err
		}
	}

	return "", nil
}

func tryGetJobIndexOnce(clientset kubernetes.Interface) (string, bool, error) {
	var (
		indexKey      = "index"
		configMapName = "runai-cli-index"
	)

	configMap, err := clientset.CoreV1().ConfigMaps(runaiNamespace).Get(configMapName, metav1.GetOptions{})

	// If configmap does not exists than cannot get a job index for the job
	if err != nil {
		return "", false, err
	}

	lastIndex, err := strconv.Atoi(configMap.Data[indexKey])

	if err != nil {
		return "", false, err
	}

	newIndex := fmt.Sprintf("%d", lastIndex+1)
	configMap.Data[indexKey] = newIndex

	_, err = clientset.CoreV1().ConfigMaps(runaiNamespace).Update(configMap)

	// Might be someone already updated this configmap. Try the process again.
	if err != nil {
		return "", true, err
	}

	return newIndex, false, nil
}

func convertOldCommandArgsFlags(cmd *cobra.Command, submitArgs *submitArgs, args []string) []string {
	commandArgs, isCommand := mergeOldCommandAndArgsWithNew(cmd.ArgsLenAtDash(), args, submitArgs.SpecCommand, submitArgs.SpecArgs, submitArgs.Command)
	if isCommand != nil && *isCommand {
		submitArgs.SpecCommand = commandArgs
		submitArgs.SpecArgs = []string{}
	} else {
		submitArgs.SpecCommand = []string{}
		submitArgs.SpecArgs = commandArgs
	}
	submitArgs.Command = isCommand
	return commandArgs
}

func mergeOldCommandAndArgsWithNew(argsLenAtDash int, positionalArgs, oldCommand, oldArgs []string, isCommand *bool) ([]string, *bool) {
	if argsLenAtDash == -1 {
		argsLenAtDash = len(positionalArgs)
	}

	argsAfterDash := positionalArgs[argsLenAtDash:]
	if len(argsAfterDash) != 0 {
		return argsAfterDash, isCommand
	}

	isAnyCommand := false
	if len(oldCommand) != 0 {
		isAnyCommand = true
	}
	return append(oldCommand, oldArgs...), &isAnyCommand
}

func getJobNameWithSuffixGenerationFlag(cmd *cobra.Command, args []string, submitArgs *submitArgs) (string, bool, error) {
	argsLenUntilDash := cmd.ArgsLenAtDash()
	argsUntilDash := args
	if argsLenUntilDash != -1 {
		argsUntilDash = args[:argsLenUntilDash]
	}
	if submitArgs.NameParameter != "" {
		if len(argsUntilDash) > 0 {
			return "", false, fmt.Errorf("unexpected arguments %v", argsUntilDash)
		}
		return submitArgs.NameParameter, false, nil
	} else if len(argsUntilDash) > 0 {
		//TODO: Show the user that the positional argument is deprecated once we feel confortable to tell it the user
		//log.Info("Submitting the job name as a positional argument has been deprecated, please use --name flag instead")
		return argsUntilDash[0], false, nil
	} else if submitArgs.NamePrefix != "" {
		return submitArgs.NamePrefix, true, nil
	}
	return jobDefaultName, true, nil
}

func AlignArgsPreParsing(args []string) []string {
	if len(args) < 2 || (args[1] != submitCommand && args[1] != SubmitMpiCommand) {
		return args
	}

	dashIndex := -1
	for i, arg := range args {
		if arg == dashArg {
			dashIndex = i
		}
	}

	if dashIndex == -1 {
		for i, arg := range args {
			if arg == fmt.Sprintf("%s%s", dashArg, commandFlag) {
				log.Info(fmt.Sprintf("using %s%s as string flag has been deprecated. Please see usage information", dashArg, commandFlag))
				args[i] = fmt.Sprintf("%s%s", dashArg, oldCommandFlag)
			}
		}
	}
	return args
}
