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
	"os"
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

	AlwaysPullImage    *bool    `yaml:"alwaysPullImage,omitempty"`
	Volumes            []string `yaml:"volume,omitempty"`
	WorkingDir         string   `yaml:"workingDir,omitempty"`
	RunAsUser          string   `yaml:"runAsUser,omitempty"`
	RunAsGroup         string   `yaml:"runAsGroup,omitempty"`
	SupplementalGroups []int    `yaml:"supplementalGroups,omitempty"`
	RunAsCurrentUser   bool
	Command            []string          `yaml:"command"`
	LocalImage         *bool             `yaml:"localImage,omitempty"`
	LargeShm           *bool             `yaml:"shm,omitempty"`
	Ports              []string          `yaml:"ports,omitempty"`
	Labels             map[string]string `yaml:"labels,omitempty"`
	HostIPC            *bool             `yaml:"hostIPC,omitempty"`
	HostNetwork        *bool             `yaml:"hostNetwork,omitempty"`
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

	// if s.DataDir == "" {
	// 	return fmt.Errorf("--dataDir must be set")
	// }

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

func (submitArgs *submitArgs) addCommonFlags(command *cobra.Command) {
	var defaultUser string
	currentUser, err := user.Current()
	if err != nil {
		defaultUser = ""
	} else {
		defaultUser = currentUser.Username
	}

	command.Flags().StringVar(&nameParameter, "name", "", "Job name")
	command.Flags().MarkDeprecated("name", "please use positional argument instead")

	flags.AddFloat64NullableFlagP(command.Flags(), &(submitArgs.GPU), "gpu", "g", "Number of GPUs to allocate to the Job.")
	flags.AddBoolNullableFlag(command.Flags(), &(submitArgs.Interactive), "interactive", "Mark this Job as interactive.")
	command.Flags().StringVar(&(submitArgs.CPU), "cpu", "", "CPU units to allocate for the job (0.5, 1)")
	command.Flags().StringVar(&(submitArgs.Memory), "memory", "", "CPU Memory to allocate for this job (1G, 20M)")
	command.Flags().StringVar(&(submitArgs.CPULimit), "cpu-limit", "", "CPU limit for the job (0.5, 1)")
	command.Flags().StringVar(&(submitArgs.MemoryLimit), "memory-limit", "", "Memory limit for this job (1G, 20M)")
	command.Flags().StringVarP(&(submitArgs.Project), "project", "p", "", "Specifies the project to which the command applies. By default, commands apply to the default project. To change the default project use 'runai project set <project name>'.")
	command.Flags().StringVarP(&(submitArgs.User), "user", "u", defaultUser, "Use different user to run the Job.")
	command.Flags().StringVarP(&(submitArgs.Image), "image", "i", "", "Container image to use when creating the job.")
	command.Flags().StringArrayVar(&(submitArgs.Args), "args", []string{}, "Arguments to pass to the command run on container start. Use together with --command.")
	command.Flags().StringVar(&(submitArgs.NodeType), "node-type", "", "Enforce node type affinity by setting a node-type label.")
	command.Flags().StringArrayVarP(&(submitArgs.EnvironmentVariable), "environment", "e", []string{}, "Set environment variables in the container.")
	command.Flags().MarkHidden("user")
	// Will not submit the job to the cluster, just print the template to the screen
	command.Flags().BoolVar(&dryRun, "dry-run", false, "run as dry run")
	command.Flags().MarkHidden("dry-run")

	flags.AddBoolNullableFlag(command.Flags(), &(submitArgs.AlwaysPullImage), "always-pull-image", "Always pull latest version of the image.")
	command.Flags().StringArrayVarP(&(submitArgs.Volumes), "volume", "v", []string{}, "Volumes to mount into the container.")
	command.Flags().StringArrayVar(&(submitArgs.Volumes), "volumes", []string{}, "Volumes to mount into the container.")
	command.Flags().MarkDeprecated("volumes", "please use 'volume' flag instead.")
	command.Flags().StringVar(&(submitArgs.WorkingDir), "working-dir", "", "Set the container's working directory.")
	command.Flags().StringArrayVar(&(submitArgs.Command), "command", []string{}, "Run this command on container start. Use together with --args.")
	command.Flags().BoolVar(&(submitArgs.RunAsCurrentUser), "run-as-user", false, "Run the job container in the context of the current user of the Run:AI CLI rather than the root user.")
	flags.AddBoolNullableFlag(command.Flags(), &(submitArgs.LocalImage), "local-image", "Use an image stored locally on the machine running the job.")
	flags.AddBoolNullableFlag(command.Flags(), &(submitArgs.LargeShm), "large-shm", "Mount a large /dev/shm device.")
	command.Flags().StringArrayVar(&(submitArgs.Ports), "port", []string{}, "Expose ports from the Job container.")
	command.Flags().StringVarP(&(configArg), "template", "t", "", "Use a specific template to run this job (otherwise use the default template if exists).")
	flags.AddBoolNullableFlag(command.Flags(), &(submitArgs.HostIPC), "host-ipc", "Use the host's ipc namespace.")
	flags.AddBoolNullableFlag(command.Flags(), &(submitArgs.HostNetwork), "host-network", "Use the host's network stack inside the container.")
}

func (submitArgs *submitArgs) setCommonRun(cmd *cobra.Command, args []string, kubeClient *client.Client, clientset kubernetes.Interface, configValues *string) {
	util.SetLogLevel(logLevel)
	if nameParameter == "" && len(args) >= 1 {
		name = args[0]
	} else {
		name = nameParameter
	}

	if name == "" {
		cmd.Help()
		fmt.Println("")
		fmt.Println("Name must be provided for the job.")
		os.Exit(1)
	}

	var errs = validation.IsDNS1035Label(name)
	if len(errs) > 0 {
		fmt.Println("")
		fmt.Println("Job names must consist of lower case alphanumeric characters or '-' and start with an alphabetic character (e.g. 'my-name',  or 'abc-123')")
		os.Exit(1)
	}

	submitArgs.Name = name

	namespaceInfo, err := flags.GetNamespaceToUseFromProjectFlagAndPrintError(cmd, kubeClient)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if namespaceInfo.ProjectName == "" {
		os.Exit(1)
	}

	clusterConfig, err := clusterConfig.GetClusterConfig(clientset)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	submitArgs.Namespace = namespaceInfo.Namespace
	submitArgs.Project = namespaceInfo.ProjectName
	if clusterConfig.EnforceRunAsUser || submitArgs.RunAsCurrentUser {
		currentUser, err := user.Current()
		if err == nil {
			groups, err := syscall.Getgroups()
			if err == nil {
				submitArgs.SupplementalGroups = groups
			} else {
				log.Debugf("Could not retrieve list of groups for user: %s", err.Error())
			}
			submitArgs.RunAsUser = currentUser.Uid
			submitArgs.RunAsGroup = currentUser.Gid
		}
	}

	index, err := getJobIndex(clientset)

	if err != nil {
		log.Debug("Could not get job index. Will not set a label.")
	} else {
		submitArgs.Labels = make(map[string]string)
		submitArgs.Labels["runai/job-index"] = index
	}

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
			fmt.Printf("Could not find runai template %s. Please run '%s template list'", configArg, config.CLIName)
			os.Exit(1)
		}
	}

	if configToUse != nil {
		*configValues = configToUse.Values
	}
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
