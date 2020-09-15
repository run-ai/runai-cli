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

package commands

import (
	"fmt"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/kubeflow/arena/cmd/arena/commands/flags"
	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/clusterConfig"
	"github.com/kubeflow/arena/pkg/config"
	"github.com/kubeflow/arena/pkg/templates"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
)

const (
	runaiNamespace = "runai"
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
	NodeType            string   `yaml:"node_type,omitempty"`
	Args                []string `yaml:"args,omitempty"`
	CPU                 string   `yaml:"cpu,omitempty"`
	CPULimit            string   `yaml:"cpuLimit,omitempty"`
	Memory              string   `yaml:"memory,omitempty"`
	MemoryLimit         string   `yaml:"memoryLimit,omitempty"`
	EnvironmentVariable []string `yaml:"environment,omitempty"`

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
	StdIn                      *bool              `yaml:"stdin,omitempty"`
	TTY                        *bool              `yaml:"tty,omitempty"`
	Attach                     *bool              `yaml:"attach,omitempty"`
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
	if name == "" {
		return fmt.Errorf("--name must be set")
	}

	// return fmt.Errorf("must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.")
	err := util.ValidateJobName(name)
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

func (submitArgs *submitArgs) addCommonFlags(cmd *cobra.Command) {
	var defaultUser string
	currentUser, err := user.Current()
	if err != nil {
		defaultUser = ""
	} else {
		defaultUser = currentUser.Username
	}

	cmd.Flags().StringVar(&nameParameter, "name", "", "Job name")
	cmd.Flags().MarkDeprecated("name", "please use positional argument instead")

	flags.AddFloat64NullableFlagP(cmd.Flags(), &(submitArgs.GPU), "gpu", "g", "Number of GPUs to allocate to the Job.")
	flags.AddBoolNullableFlag(cmd.Flags(), &(submitArgs.Interactive), "interactive", "Mark this Job as interactive.")
	cmd.Flags().StringVar(&(submitArgs.CPU), "cpu", "", "CPU units to allocate for the job (0.5, 1)")
	cmd.Flags().StringVar(&(submitArgs.Memory), "memory", "", "CPU Memory to allocate for this job (1G, 20M)")
	cmd.Flags().StringVar(&(submitArgs.CPULimit), "cpu-limit", "", "CPU limit for the job (0.5, 1)")
	cmd.Flags().StringVar(&(submitArgs.MemoryLimit), "memory-limit", "", "Memory limit for this job (1G, 20M)")
	cmd.Flags().StringVarP(&(submitArgs.Project), "project", "p", "", "Specifies the project to which the command applies. By default, commands apply to the default project. To change the default project use 'runai project set <project name>'.")
	cmd.Flags().StringVarP(&(submitArgs.User), "user", "u", defaultUser, "Use different user to run the Job.")
	cmd.Flags().StringVarP(&(submitArgs.Image), "image", "i", "", "Container image to use when creating the job.")
	cmd.Flags().StringArrayVar(&(submitArgs.Args), "args", []string{}, "Arguments to pass to the command run on container start. Use together with --command.")
	cmd.Flags().StringVar(&(submitArgs.NodeType), "node-type", "", "Enforce node type affinity by setting a node-type label.")
	cmd.Flags().StringArrayVarP(&(submitArgs.EnvironmentVariable), "environment", "e", []string{}, "Set environment variables in the container.")
	cmd.Flags().MarkHidden("user")
	// Will not submit the job to the cluster, just print the template to the screen
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run as dry run")
	cmd.Flags().MarkHidden("dry-run")

	flags.AddBoolNullableFlag(cmd.Flags(), &(submitArgs.AlwaysPullImage), "always-pull-image", "Always pull latest version of the image.")
	cmd.Flags().StringArrayVarP(&(submitArgs.Volumes), "volume", "v", []string{}, "Volumes to mount into the container.")
	cmd.Flags().StringArrayVar(&(submitArgs.PersistentVolumes), "pvc", []string{}, "Kubernetes provisioned persistent volumes to mount into the container. Directives are given in the form 'StorageClass[optional]:Size:ContainerMountPath[optional]:ro[optional]")
	cmd.Flags().StringArrayVar(&(submitArgs.Volumes), "volumes", []string{}, "Volumes to mount into the container.")
	cmd.Flags().MarkDeprecated("volumes", "please use 'volume' flag instead.")
	cmd.Flags().StringVar(&(submitArgs.WorkingDir), "working-dir", "", "Set the container's working directory.")
	cmd.Flags().StringArrayVar(&(submitArgs.Command), "command", []string{}, "Run this command on container start. Use together with --args.")
	cmd.Flags().BoolVar(&(submitArgs.RunAsCurrentUser), "run-as-user", false, "Run the job container in the context of the current user of the Run:AI CLI rather than the root user.")
	flags.AddBoolNullableFlag(cmd.Flags(), &(submitArgs.CreateHomeDir), "create-user-dir", "Create a temporary home directory for the user in the container.  Data saved in this directory will not be saved when the container exits. The flag is set by default to true when the --run-as-user flag is used, and false if not.")

	cmd.Flags().BoolVarP(submitArgs.TTY, "tty", "t", false, "Allocate a TTY for the container.")
	flags.AddBoolNullableFlag(cmd.Flags(), &submitArgs.StdIn, "stdin", "Keep stdin open on the container(s) in the pod, even if nothing is attached.")
	flags.AddBoolNullableFlag(cmd.Flags(), &submitArgs.Attach, "attach", `If true, wait for the Pod to start running, and then attach to the Pod as if 'runai attach ...' were called. Default false, unless '--stdin' is set, in which case the default is true.`)
	cmd.Flags().BoolVar(&(submitArgs.PreventPrivilegeEscalation), "prevent-privilege-escalation", false, "Prevent the job’s container from gaining additional privileges after start.")
	flags.AddBoolNullableFlag(cmd.Flags(), &submitArgs.LocalImage, "local-image", "Use an image stored locally on the machine running the job.")
	flags.AddBoolNullableFlag(cmd.Flags(), &submitArgs.LargeShm, "large-shm", "Mount a large /dev/shm device.")
	cmd.Flags().StringArrayVar(&(submitArgs.Ports), "port", []string{}, "Expose ports from the Job container.")
	cmd.Flags().StringVarP(&(configArg), "template", "", "", "Use a specific template to run this job (otherwise use the default template if exists).")
	flags.AddBoolNullableFlag(cmd.Flags(), &(submitArgs.HostIPC), "host-ipc", "Use the host's ipc namespace.")
	flags.AddBoolNullableFlag(cmd.Flags(), &(submitArgs.HostNetwork), "host-network", "Use the host's network stack inside the container.")
}

func (submitArgs *submitArgs) setCommonRun(cmd *cobra.Command, args []string, kubeClient *client.Client, clientset kubernetes.Interface, configValues *string) error {
	util.SetLogLevel(logLevel)
	if nameParameter == "" && len(args) >= 1 {
		name = args[0]
	} else {
		name = nameParameter
	}

	if name == "" {
		cmd.Help()
		fmt.Println("")
		return fmt.Errorf("Name must be provided for the job.")
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
		if submitArgs.CreateHomeDir == nil {
			t := true
			submitArgs.CreateHomeDir = &t
		}
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

	if IsBoolPTrue(submitArgs.TTY) && !IsBoolPTrue(submitArgs.StdIn) {
		return fmt.Errorf("--stdin is required for containers with -t/--tty=true")
	}

	// by default when the user set --attach the --stdin and --tty set to true
	if IsBoolPTrue(submitArgs.Attach) {
		if submitArgs.StdIn == nil {
			submitArgs.StdIn = BoolP(true)
		}

		if submitArgs.TTY == nil {
			submitArgs.TTY = BoolP(true)
		}
	// by default when the user set --stdin the --attach set to true
	} else if IsBoolPTrue(submitArgs.StdIn) && submitArgs.Attach == nil {
		submitArgs.Attach = BoolP(true)
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


func BoolP(b bool) *bool {
	return &b
}

func IsBoolPTrue(b *bool) bool {
	return b != nil && *b
}