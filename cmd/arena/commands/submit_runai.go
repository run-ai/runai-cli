package commands

import (
	"fmt"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kubeflow/arena/cmd/arena/commands/flags"
	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/clusterConfig"
	"github.com/kubeflow/arena/pkg/config"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	runaiChart       string
	ttlAfterFinished *time.Duration
	configArg        string
)

const (
	defaultRunaiTrainingType = "runai"
	runaiNamespace           = "runai"
)

func NewRunaiJobCommand() *cobra.Command {

	submitArgs := NewSubmitRunaiJobArgs()
	var command = &cobra.Command{
		Use:   "submit [NAME]",
		Short: "Submit a new job.",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			chartsFolder, err := util.GetChartsFolder()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			runaiChart = path.Join(chartsFolder, "runai")

			kubeClient, err := client.GetClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			clientset := kubeClient.GetClientset()

			submitArgs.setCommonRun(cmd, args, kubeClient)

			index, err := getJobIndex(clientset)

			if err != nil {
				log.Debug("Could not get job index. Will not set a label.")
			} else {
				submitArgs.Labels = make(map[string]string)
				submitArgs.Labels["runai/job-index"] = index
			}

			if ttlAfterFinished != nil {
				ttlSeconds := int(math.Round(ttlAfterFinished.Seconds()))
				log.Debugf("Using time to live seconds %d", ttlSeconds)
				submitArgs.TTL = &ttlSeconds
			}

			if submitArgs.IsJupyter {
				submitArgs.UseJupyterDefaultValues()
			}

			err = submitRunaiJob(args, submitArgs, clientset)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			printJobInfoIfNeeded(submitArgs)
			if submitArgs.IsJupyter || (submitArgs.Interactive != nil && *submitArgs.Interactive && submitArgs.ServiceType == "portforward") {
				err = kubectl.WaitForReadyStatefulSet(submitArgs.Name, submitArgs.Namespace)

				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				if submitArgs.IsJupyter {
					runaiTrainer := NewRunaiTrainer(*kubeClient)
					job, err := runaiTrainer.GetTrainingJob(submitArgs.Name, submitArgs.Namespace)

					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					pod := job.ChiefPod()
					logs, err := kubectl.Logs(pod.Name, pod.Namespace)

					token, err := getTokenFromJupyterLogs(string(logs))

					if err != nil {
						fmt.Println(err)
						fmt.Printf("Please run '%s logs %s' to view the logs.\n", config.CLIName, submitArgs.Name)
					}

					fmt.Printf("Jupyter notebook token: %s\n", token)
				}

				if submitArgs.Interactive != nil && *submitArgs.Interactive && submitArgs.ServiceType == "portforward" {
					localPorts := []string{}
					for _, port := range submitArgs.Ports {
						split := strings.Split(port, ":")
						localPorts = append(localPorts, split[0])
					}

					localUrls := []string{}
					for _, localPort := range localPorts {
						localUrls = append(localUrls, fmt.Sprintf("localhost:%s", localPort))
					}

					accessPoints := strings.Join(localUrls, ",")
					fmt.Printf("Open access point(s) to service from %s\n", accessPoints)
					err = kubectl.PortForward(localPorts, submitArgs.Name, submitArgs.Namespace)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
				}
			}
		},
	}

	submitArgs.addCommonFlags(command)
	submitArgs.addFlags(command)

	return command
}

func printJobInfoIfNeeded(submitArgs *submitRunaiJobArgs) {
	if submitArgs.Interactive != nil && *submitArgs.Interactive && submitArgs.IsPreemptible != nil && *submitArgs.IsPreemptible {
		fmt.Println("Warning: Using the preemptible flag may lead to your resources being preempted without notice")
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

func getTokenFromJupyterLogs(logs string) (string, error) {
	re, err := regexp.Compile(`\?token=(.*)\n`)
	if err != nil {
		return "", err
	}

	res := re.FindStringSubmatch(logs)
	if len(res) < 2 {
		return "", fmt.Errorf("Could not find token string in logs")
	}
	return res[1], nil
}

func NewSubmitRunaiJobArgs() *submitRunaiJobArgs {
	return &submitRunaiJobArgs{}
}

type submitRunaiJobArgs struct {
	// These arguments should be omitted when empty, to support default values file created in the cluster
	// So any empty ones won't override the default values
	submitArgs       `yaml:",inline"`
	HostIPC          *bool             `yaml:"hostIPC,omitempty"`
	GPUInt           *int              `yaml:"gpuInt,omitempty"`
	GPUFraction      string            `yaml:"gpuFraction,omitempty"`
	GPUFractionFixed string            `yaml:"gpuFractionFixed,omitempty"`
	Command          []string          `yaml:"command"`
	Ports            []string          `yaml:"ports,omitempty"`
	ServiceType      string            `yaml:"serviceType,omitempty"`
	Elastic          *bool             `yaml:"elastic,omitempty"`
	LargeShm         *bool             `yaml:"shm,omitempty"`
	NumberProcesses  int               `yaml:"numProcesses"` // --workers
	LocalImage       *bool             `yaml:"localImage,omitempty"`
	HostNetwork      *bool             `yaml:"hostNetwork,omitempty"`
	TTL              *int              `yaml:"ttlSecondsAfterFinished,omitempty"`
	Completions      *int              `yaml:"completions,omitempty"`
	Parallelism      *int              `yaml:"parallelism,omitempty"`
	Labels           map[string]string `yaml:"labels,omitempty"`
	IsJupyter        bool
	IsPreemptible    *bool `yaml:"isPreemptible,omitempty"`
}

func (sa *submitRunaiJobArgs) UseJupyterDefaultValues() {
	var (
		jupyterPort        = "8888"
		jupyterImage       = "jupyter/scipy-notebook"
		jupyterCommand     = "start-notebook.sh"
		jupyterArgs        = "--NotebookApp.base_url=/%s"
		jupyterServiceType = "portforward"
	)

	interactive := true
	sa.Interactive = &interactive
	if len(sa.Ports) == 0 {
		sa.Ports = []string{jupyterPort}
		log.Infof("Exposing default jupyter notebook port %s", jupyterPort)
	}
	if sa.Image == "" {
		sa.Image = "jupyter/scipy-notebook"
		log.Infof("Using default jupyter notebook image \"%s\"", jupyterImage)
	}
	if sa.ServiceType == "" {
		sa.ServiceType = jupyterServiceType
		log.Infof("Using default jupyter notebook service type %s", jupyterServiceType)
	}
	if len(sa.Command) == 0 && sa.ServiceType == "ingress" {
		sa.Command = []string{jupyterCommand}
		log.Infof("Using default jupyter notebook command for using ingress service \"%s\"", jupyterCommand)
	}
	if len(sa.Args) == 0 && sa.ServiceType == "ingress" {
		baseUrlArg := fmt.Sprintf(jupyterArgs, name)
		sa.Args = []string{baseUrlArg}
		log.Infof("Using default jupyter notebook command argument for using ingress service \"%s\"", baseUrlArg)
	}
}

// add flags to submit spark args
func (sa *submitRunaiJobArgs) addFlags(command *cobra.Command) {

	flags.AddBoolNullableFlag(command.Flags(), &(sa.HostIPC), "host-ipc", "Use the host's ipc namespace.")
	command.Flags().StringArrayVar(&(sa.Ports), "port", []string{}, "Expose ports from the Job container.")
	command.Flags().StringVarP(&(sa.ServiceType), "service-type", "s", "", "Specify service exposure for interactive jobs. Options are: portforward, loadbalancer, nodeport, ingress.")
	command.Flags().BoolVar(&(sa.IsJupyter), "jupyter", false, "Shortcut for running a jupyter notebook using a pre-created image and a default notebook configuration.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.Elastic), "elastic", "Mark the job as elastic.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.IsPreemptible), "preemptible", "Mark an interactive job as preemptible. Preemptible jobs can be scheduled above guaranteed quota but may be reclaimed at any time.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.LargeShm), "large-shm", "Mount a large /dev/shm device.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.LocalImage), "local-image", "Use an image stored locally on the machine running the job.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.HostNetwork), "host-network", "Use the host's network stack inside the container.")
	flags.AddIntNullableFlag(command.Flags(), &(sa.Completions), "completions", "The number of successful pods required for this job to be completed.")
	flags.AddIntNullableFlag(command.Flags(), &(sa.Parallelism), "parallelism", "The number of pods this job tries to run in parallel at any instant.")
	command.Flags().MarkHidden("parallelism")
	command.Flags().MarkHidden("completions")
	command.Flags().StringArrayVar(&(sa.Command), "command", []string{}, "Run this command on container start. Use together with --args.")
	command.Flags().BoolVar(&(sa.RunAsCurrentUser), "run-as-user", false, "Run the job container in the context of the current user of the Run:AI CLI rather than the root user.")

	flags.AddDurationNullableFlagP(command.Flags(), &(ttlAfterFinished), "ttl-after-finish", "", "Define the duration, post job finish, after which the job is automatically deleted (e.g. 5s, 2m, 3h).")

	command.Flags().StringVarP(&(configArg), "template", "t", "", "Use a specific template to run this job (otherwise use the default template if exists).")
}

func submitRunaiJob(args []string, submitArgs *submitRunaiJobArgs, clientset kubernetes.Interface) error {
	configs := clusterConfig.NewClusterConfigs(clientset)

	var configToUse *clusterConfig.ClusterConfig
	var err error
	if configArg == "" {
		configToUse, err = configs.GetClusterDefaultConfig()
	} else {
		configToUse, err = configs.GetClusterConfig(configArg)
		if configToUse == nil {
			return fmt.Errorf("Could not find runai template %s. Please run '%s template list'", configArg, config.CLIName)
		}
	}

	if submitArgs.Completions == nil && submitArgs.Parallelism != nil {
		// Setting parallelism without setting completions causes kubernetes to treat this job as having a work queue. For more info: https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/#job-patterns
		return fmt.Errorf("if the parallelism flag is set, you must also set the number of successful pod completions required for this job to complete (use --completions <number_of_required_completions>)")
	}

	if err != nil {
		return err
	}

	configValues := ""
	if configToUse != nil {
		configValues = configToUse.Values
	}

	err = handleRequestedGPUs(submitArgs)
	if err != nil {
		return err
	}

	err = workflow.SubmitJob(submitArgs.Name, defaultRunaiTrainingType, submitArgs.Namespace, submitArgs, configValues, runaiChart, clientset, dryRun)
	if err != nil {
		return err
	}

	fmt.Printf("The job '%s' has been submitted successfully\n", submitArgs.Name)
	fmt.Printf("You can run `%s get %s -p %s` to check the job status\n", config.CLIName, submitArgs.Name, submitArgs.Project)
	return nil
}
