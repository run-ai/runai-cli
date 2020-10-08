package submit

import (
	"fmt"
	"math"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/trainer"
	"github.com/run-ai/runai-cli/cmd/attach"


	runaiclientset "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned"
	raUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/config"
	"github.com/run-ai/runai-cli/pkg/util"
	"github.com/run-ai/runai-cli/pkg/util/kubectl"
	"github.com/run-ai/runai-cli/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	runaiChart       string
	ttlAfterFinished *time.Duration
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
			configValues := ""
			runaijobClient := runaiclientset.NewForConfigOrDie(kubeClient.GetRestConfig())

			err = submitArgs.setCommonRun(cmd, args, kubeClient, clientset, &configValues)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if ttlAfterFinished != nil {
				ttlSeconds := int(math.Round(ttlAfterFinished.Seconds()))
				log.Debugf("Using time to live seconds %d", ttlSeconds)
				submitArgs.TTL = &ttlSeconds
			}

			if submitArgs.IsJupyter {
				submitArgs.UseJupyterDefaultValues()
			}

			err = submitRunaiJob(args, submitArgs, clientset, *runaijobClient, &configValues)
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
					runaiTrainer := trainer.NewRunaiTrainer(*kubeClient)
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

			if submitArgs.Attach != nil && *submitArgs.Attach {
				if err := attach.Attach(cmd, submitArgs.Name, raUtil.IsBoolPTrue(submitArgs.StdIn), raUtil.IsBoolPTrue(submitArgs.TTY), "", attach.DefaultAttachTimeout); err != nil {
					fmt.Println(err)
					os.Exit(1)
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
	GPUFractionFixed string `yaml:"gpuFractionFixed,omitempty"`
	ServiceType      string `yaml:"serviceType,omitempty"`
	Elastic          *bool  `yaml:"elastic,omitempty"`
	NumberProcesses  int    `yaml:"numProcesses"` // --workers
	TTL              *int   `yaml:"ttlSecondsAfterFinished,omitempty"`
	Completions      *int   `yaml:"completions,omitempty"`
	Parallelism      *int   `yaml:"parallelism,omitempty"`
	BackoffLimit     *int   `yaml:"backoffLimit,omitempty"`
	IsJupyter        bool
	IsPreemptible    *bool `yaml:"isPreemptible,omitempty"`
	IsRunaiJob       *bool `yaml:"isRunaiJob,omitempty"`
}

func (sa *submitRunaiJobArgs) UseJupyterDefaultValues() {
	var (
		jupyterPort        = "8888"
		jupyterImage       = "jupyter/scipy-notebook"
		jupyterCommand     = "start-notebook.sh"
		jupyterArgs        = "--NotebookApp.base_url=/%s-%s"
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
		baseUrlArg := fmt.Sprintf(jupyterArgs, sa.Project, sa.Name)
		sa.Args = []string{baseUrlArg}
		log.Infof("Using default jupyter notebook command argument for using ingress service \"%s\"", baseUrlArg)
	}
}

// add flags to submit spark args
func (sa *submitRunaiJobArgs) addFlags(command *cobra.Command) {

	command.Flags().StringVarP(&(sa.ServiceType), "service-type", "s", "", "Specify service exposure for interactive jobs. Options are: portforward, loadbalancer, nodeport, ingress.")
	command.Flags().BoolVar(&(sa.IsJupyter), "jupyter", false, "Shortcut for running a jupyter notebook using a pre-created image and a default notebook configuration.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.Elastic), "elastic", "", "Mark the job as elastic.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.IsPreemptible), "preemptible", "", "Mark an interactive job as preemptible. Preemptible jobs can be scheduled above guaranteed quota but may be reclaimed at any time.")
	flags.AddIntNullableFlag(command.Flags(), &(sa.Completions), "completions", "The number of successful pods required for this job to be completed. Used for Hyperparameter optimization.")
	flags.AddIntNullableFlag(command.Flags(), &(sa.Parallelism), "parallelism", "The number of pods this job tries to run in parallel at any time.  Used for Hyperparameter optimization.")
	flags.AddIntNullableFlag(command.Flags(), &(sa.BackoffLimit), "backoffLimit", "The number of times the job will be retried before failing. Default 6.")
	flags.AddDurationNullableFlagP(command.Flags(), &(ttlAfterFinished), "ttl-after-finish", "", "Define the duration, post job finish, after which the job is automatically deleted (e.g. 5s, 2m, 3h).")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.IsRunaiJob), "runai-job", "", "submit a job of resource runaijob")
	command.Flags().MarkHidden("runai-job")
}

func submitRunaiJob(args []string, submitArgs *submitRunaiJobArgs, clientset kubernetes.Interface, runaiclientset runaiclientset.Clientset, configValues *string) error {
	err2 := verifyHPOFlags(submitArgs)
	if err2 != nil {
		return err2
	}

	handleRunaiJobCRD(submitArgs, runaiclientset)
	submitArgs.Labels[kubectl.BaseNameLabel] = submitArgs.Name
	mutatedName, err := workflow.SubmitJob(submitArgs.Name, trainer.DefaultRunaiTrainingType, submitArgs.Namespace, submitArgs, *configValues, runaiChart, clientset, dryRun)
	if err != nil {
		return err
	}
	submitArgs.Name = mutatedName

	fmt.Printf("The job '%s' has been submitted successfully\n", submitArgs.Name)
	fmt.Printf("You can run `%s get %s -p %s` to check the job status\n", config.CLIName, submitArgs.Name, submitArgs.Project)
	return nil
}

// For backward compatibility - remove once all customers have runaijob crd
func handleRunaiJobCRD(submitArgs *submitRunaiJobArgs, runaiclientset runaiclientset.Clientset) {
	isRunaiJob := true
	submitArgs.IsRunaiJob = &isRunaiJob
	_, err := runaiclientset.RunV1().RunaiJobs("").List(metav1.ListOptions{})
	if err != nil {
		*submitArgs.IsRunaiJob = false
	}
}

func verifyHPOFlags(submitArgs *submitRunaiJobArgs) error {
	if submitArgs.Parallelism != nil && *submitArgs.Parallelism > 1 {
		if submitArgs.Completions == nil {
			// Setting parallelism without setting completions causes kubernetes to treat this job as having a work queue. For more info: https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/#job-patterns
			return fmt.Errorf("if the parallelism flag is set, you must also set the number of successful pod completions required for this job to complete (use --completions <number_of_required_completions>)")
		}
		if submitArgs.Elastic != nil {
			return fmt.Errorf("elasitc jobs can't run with Parallelism")
		}
		if submitArgs.Interactive != nil {
			return fmt.Errorf("interactive jobs can't run with Parallelism")
		}
		if *submitArgs.GPU != 0 && *submitArgs.GPU < 1 {
			return fmt.Errorf("usage of GPU Fractions in conjunction with parallelism is not supported")
		}
	}
	return nil
}
