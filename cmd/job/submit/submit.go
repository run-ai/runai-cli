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
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/global"
	raUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/clusterConfig"
	"github.com/run-ai/runai-cli/pkg/config"
	"github.com/run-ai/runai-cli/pkg/templates"
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
	nameParameter string
	dryRun        bool

	envs        []string
	selectors   []string
	tolerations []string
	dataset     []string
	dataDirs    []string
	annotations []string
	configArg   string
)

// The common parts of the submitAthd
type submitArgs struct {
	// Name       string   `yaml:"name"`       // --name
	NodeSelectors map[string]string `yaml:"nodeSelectors"` // --selector
	Tolerations   []string          `yaml:"tolerations"`   // --toleration
	Image         string            `yaml:"image"`         // --image
	Envs          map[string]string `yaml:"envs"`          // --envs
	// for horovod
	Mode string `yaml:"mode"`
	// --mode
	// SSHPort     int               `yaml:"sshPort"`  // --sshPort
	Retry int `yaml:"retry"` // --retry
	// DataDir  string            `yaml:"dataDir"`  // --dataDir
	DataSet  map[string]string `yaml:"dataset"`
	DataDirs []dataDirVolume   `yaml:"dataDirs"`

	EnableRDMA bool `yaml:"enableRDMA"` // --rdma
	UseENI     bool `yaml:"useENI"`

	Annotations map[string]string `yaml:"annotations"`

	IsNonRoot          bool                      `yaml:"isNonRoot"`
	PodSecurityContext limitedPodSecurityContext `yaml:"podSecurityContext"`
	Project            string                    `yaml:"project,omitempty"`
	Interactive        *bool                     `yaml:"interactive,omitempty"`
	User               string                    `yaml:"user,omitempty"`
	PriorityClassName  string                    `yaml:"priorityClassName"`
	// Name       string   `yaml:"name"`       // --name
	Name                string
	Namespace           string
	GPU                 *float64 `yaml:"gpu,omitempty"`
	GPUInt              *int     `yaml:"gpuInt,omitempty"`
	GPUFraction         string   `yaml:"gpuFraction,omitempty"`
	NodeType            string   `yaml:"node_type,omitempty"`
	Args                []string `yaml:"args,omitempty"`
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
	PreventPrivilegeEscalation bool     `yaml:"preventPrivilegeEscalation"`
	CreateHomeDir              *bool    `yaml:"createHomeDir,omitempty"`
	RunAsUser                  string   `yaml:"runAsUser,omitempty"`
	RunAsGroup                 string   `yaml:"runAsGroup,omitempty"`
	SupplementalGroups         []int    `yaml:"supplementalGroups,omitempty"`
	RunAsCurrentUser           bool
	Command                    []string          `yaml:"command"`
	LocalImage                 *bool             `yaml:"localImage,omitempty"`
	LargeShm                   *bool             `yaml:"shm,omitempty"`
	Ports                      []string          `yaml:"ports,omitempty"`
	Labels                     map[string]string `yaml:"labels,omitempty"`
	HostIPC                    *bool             `yaml:"hostIPC,omitempty"`
	HostNetwork                *bool             `yaml:"hostNetwork,omitempty"`
	StdIn                      *bool             `yaml:"stdin,omitempty"`
	TTY                        *bool             `yaml:"tty,omitempty"`
	Attach                     *bool             `yaml:"attach,omitempty"`
	namePrefix				   string			 `yaml:"namePrefix,omitempty"`
	generateSuffix 			   bool
}

type dataDirVolume struct {
	HostPath      string `yaml:"hostPath"`
	ContainerPath string `yaml:"containerPath"`
	Name          string `yaml:"name"`
}

type limitedPodSecurityContext struct {
	RunAsUser          int64   `yaml:"runAsUser"`
	RunAsNonRoot       bool    `yaml:"runAsNonRoot"`
	RunAsGroup         int64   `yaml:"runAsGroup"`
	SupplementalGroups []int64 `yaml:"supplementalGroups"`
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

	if s.PriorityClassName != "" {
		err = util.ValidatePriorityClassName(s.PriorityClassName)
		if err != nil {
			return err
		}
	}

	return nil
}

// get node selectors
func (submitArgs *submitArgs) addNodeSelectors() {
	log.Debugf("node selectors: %v", selectors)
	if len(selectors) == 0 {
		submitArgs.NodeSelectors = map[string]string{}
		return
	}
	submitArgs.NodeSelectors = transformSliceToMap(selectors, "=")
}

// get tolerations labels
func (submitArgs *submitArgs) addTolerations() {
	log.Debugf("tolerations: %v", tolerations)
	if len(tolerations) == 0 {
		submitArgs.Tolerations = []string{}
		return
	}
	submitArgs.Tolerations = []string{}
	for _, taintKey := range tolerations {
		if taintKey == "all" {
			submitArgs.Tolerations = []string{"all"}
			return
		}
		submitArgs.Tolerations = append(submitArgs.Tolerations, taintKey)
	}
}

func (submitArgs *submitArgs) addCommonFlags(fbg flags.FlagsByGroups) {
	var defaultUser string
	currentUser, err := user.Current()
	if err != nil {
		defaultUser = ""
	} else {
		defaultUser = currentUser.Username
	}

	flagSet := fbg.GetOrAddFlagSet(AliasesAndShortcutsFlagGroup)
	flagSet.StringVar(&nameParameter, "name", "", "Job name")
	flags.AddBoolNullableFlag(flagSet, &(submitArgs.Interactive), "interactive", "", "Mark this Job as interactive.")
	flagSet.StringVarP(&(configArg), "template", "", "", "Use a specific template to run this job (otherwise use the default template if exists).")
	flagSet.StringVarP(&(submitArgs.Project), "project", "p", "", "Specifies the project to which the command applies. By default, commands apply to the default project. To change the default project use 'runai project set <project name>'.")
	// Will not submit the job to the cluster, just print the template to the screen
	flagSet.BoolVar(&dryRun, "dry-run", false, "Run as dry run")
	flagSet.MarkHidden("dry-run")
	flagSet.StringVar(&submitArgs.namePrefix, "job-name-prefix", "", "Use this value as job name and generate suffix to the job name")

	flagSet = fbg.GetOrAddFlagSet(ContainerDefinitionFlagGroup)
	flagSet.StringVar(&(submitArgs.ImagePullPolicy), "image-pull-policy", "Always", "the policy of image pull, set by default to \"Always\".")
	flags.AddBoolNullableFlag(flagSet, &(submitArgs.AlwaysPullImage), "always-pull-image", "", "Always pull latest version of the image.")
	flagSet.MarkDeprecated("always-pull-image", "please use 'image-pull-policy=Always' instead.")
	flagSet.StringArrayVar(&(submitArgs.Args), "args", []string{}, "Arguments to pass to the command run on container start. Use together with --command.")
	flagSet.StringArrayVarP(&(submitArgs.EnvironmentVariable), "environment", "e", []string{}, "Set environment variables in the container.")
	flagSet.StringVarP(&(submitArgs.Image), "image", "i", "", "Container image to use when creating the job.")
	flagSet.StringArrayVar(&(submitArgs.Command), "command", []string{}, "Run this command on container start. Use together with --args.")
	flags.AddBoolNullableFlag(flagSet, &submitArgs.LocalImage, "local-image", "", "Use an image stored locally on the machine running the job.")
	flagSet.MarkDeprecated("local-image", "please use 'image-pull-policy=Never' instead.")
	flags.AddBoolNullableFlag(flagSet, &submitArgs.TTY, "tty", "t", "Allocate a TTY for the container.")
	flags.AddBoolNullableFlag(flagSet, &submitArgs.StdIn, "stdin", "", "Keep stdin open on the container(s) in the pod, even if nothing is attached.")
	flags.AddBoolNullableFlag(flagSet, &submitArgs.Attach, "attach", "", `If true, wait for the Pod to start running, and then attach to the Pod as if 'runai attach ...' were called. Attach makes tty and stdin true by default. Default false`)
	flagSet.StringVar(&(submitArgs.WorkingDir), "working-dir", "", "Set the container's working directory.")
	flagSet.BoolVar(&(submitArgs.RunAsCurrentUser), "run-as-user", false, "Run the job container in the context of the current user of the Run:AI CLI rather than the root user.")

	flagSet = fbg.GetOrAddFlagSet(ResourceAllocationFlagGroup)
	flags.AddFloat64NullableFlagP(flagSet, &(submitArgs.GPU), "gpu", "g", "Number of GPUs to allocate to the Job.")
	flagSet.StringVar(&(submitArgs.CPU), "cpu", "", "CPU units to allocate for the job (0.5, 1)")
	flagSet.StringVar(&(submitArgs.Memory), "memory", "", "CPU Memory to allocate for this job (1G, 20M)")
	flagSet.StringVar(&(submitArgs.CPULimit), "cpu-limit", "", "CPU limit for the job (0.5, 1)")
	flagSet.StringVar(&(submitArgs.MemoryLimit), "memory-limit", "", "Memory limit for this job (1G, 20M)")
	flags.AddBoolNullableFlag(flagSet, &submitArgs.LargeShm, "large-shm", "", "Mount a large /dev/shm device.")

	flagSet = fbg.GetOrAddFlagSet(StorageFlagGroup)
	flagSet.StringArrayVarP(&(submitArgs.Volumes), "volume", "v", []string{}, "Volumes to mount into the container.")
	flagSet.StringArrayVar(&(submitArgs.PersistentVolumes), "pvc", []string{}, "Kubernetes provisioned persistent volumes to mount into the container. Directives are given in the form 'StorageClass[optional]:Size:ContainerMountPath[optional]:ro[optional]")
	flagSet.StringArrayVar(&(submitArgs.Volumes), "volumes", []string{}, "Volumes to mount into the container.")
	flagSet.MarkDeprecated("volumes", "please use 'volume' flag instead.")

	flagSet = fbg.GetOrAddFlagSet(NetworkFlagGroup)
	flags.AddBoolNullableFlag(flagSet, &(submitArgs.HostIPC), "host-ipc", "", "Use the host's ipc namespace.")
	flags.AddBoolNullableFlag(flagSet, &(submitArgs.HostNetwork), "host-network", "", "Use the host's network stack inside the container.")

	flagSet = fbg.GetOrAddFlagSet(JobLifecycleFlagGroup)

	flagSet = fbg.GetOrAddFlagSet(AccessControlFlagGroup)
	flags.AddBoolNullableFlag(flagSet, &submitArgs.CreateHomeDir, "create-home-dir", "", "Create a temporary home directory for the user in the container.  Data saved in this directory will not be saved when the container exits. The flag is set by default to true when the --run-as-user flag is used, and false if not.")
	flagSet.BoolVar(&(submitArgs.PreventPrivilegeEscalation), "prevent-privilege-escalation", false, "Prevent the job’s container from gaining additional privileges after start.")
	flagSet.StringVarP(&(submitArgs.User), "user", "u", defaultUser, "Use different user to run the Job.")
	flagSet.MarkHidden("user")

	flagSet = fbg.GetOrAddFlagSet(SchedulingFlagGroup)
	flagSet.StringVar(&(submitArgs.NodeType), "node-type", "", "Enforce node type affinity by setting a node-type label.")
}

func (submitArgs *submitArgs) setCommonRun(cmd *cobra.Command, args []string, kubeClient *client.Client, clientset kubernetes.Interface, configValues *string) error {
	util.SetLogLevel(global.LogLevel)
	var name string
	if nameParameter != "" {
		if len(args) > 0 || submitArgs.namePrefix != "" {
			submitArgs.generateSuffix = false
			log.Warn("Using the name provided with --name flag. Ignoring the other options")
		}
		name = nameParameter
	} else if len(args) > 0 {
		if submitArgs.namePrefix != "" {
			log.Warn("Ignoring --job-name-prefix flag")
			submitArgs.generateSuffix = false
		}
		log.Warn("Provide name without --name flag has been deprecated")
		name = args[0]
	} else if submitArgs.namePrefix != "" {
		name = submitArgs.namePrefix
		submitArgs.generateSuffix = true
	} else {
		name = jobDefaultName
		submitArgs.generateSuffix = true
	}

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
	if clusterConfig.EnforceRunAsUser || submitArgs.RunAsCurrentUser {
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
		submitArgs.PreventPrivilegeEscalation = true
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

	configs := templates.NewTemplates(clientset)
	var configToUse *templates.Template
	if configArg == "" {
		configToUse, err = configs.GetDefaultTemplate()
	} else {
		configToUse, err = configs.GetTemplate(configArg)
		if configToUse == nil {
			return fmt.Errorf("Could not find runai template %s. Please run '%s template list'", configArg, config.CLIName)
		}
	}

	if configToUse != nil {
		*configValues = configToUse.Values
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
	return nil
}

var (
	submitLong = `Submit a job.

Available Commands:
  tfjob,tf             Submit a TFJob.
  horovod,hj           Submit a Horovod Job.
  mpijob,mpi           Submit a MPIJob.
  standalonejob,sj     Submit a standalone Job.
  tfserving,tfserving  Submit a Serving Job.
  volcanojob,vj        Submit a VolcanoJob.
    `
)

func transformSliceToMap(sets []string, split string) (valuesMap map[string]string) {
	valuesMap = map[string]string{}
	for _, member := range sets {
		splits := strings.SplitN(member, split, 2)
		if len(splits) == 2 {
			valuesMap[splits[0]] = splits[1]
		}
	}

	return valuesMap
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
