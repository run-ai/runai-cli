package submit

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/templates"
	"math"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/run-ai/runai-cli/cmd/attach"
	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/trainer"

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

const (
	submitCommand = "submit"
	submitExamples = `
# Start a Training job.
runai submit --name train1 -i gcr.io/run-ai-demo/quickstart -g 1

# Start an interactive job.
runai submit --name build1 -i python -g 1 --interactive --attach

# Use GPU Fractions
runai submit --name frac05 -i gcr.io/run-ai-demo/quickstart -g 0.5

# Hyperparameter Optimization
runai submit --name hpo1 -i gcr.io/run-ai-demo/quickstart-hpo -g 1  \
    --parallelism 3 --completions 12 -v /nfs/john/hpo:/hpo

# Auto generate job name
runai submit -i gcr.io/run-ai-demo/quickstart -g 1
`
)

var (
	runaiChart       string
)

func NewRunaiJobCommand() *cobra.Command {

	submitArgs := NewSubmitRunaiJobArgs()
	var command = &cobra.Command{
		Use:     "submit [flags] -- [COMMAND] [args...] [options]",
		DisableFlagsInUseLine: true,
		Short:   "Submit a new job.",
		Example: submitExamples,
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
			runaijobClient := runaiclientset.NewForConfigOrDie(kubeClient.GetRestConfig())

			commandArgs, isCommand := convertOldCommandArgsFlags(cmd.ArgsLenAtDash(), args, submitArgs.SpecCommand, submitArgs.SpecArgs, raUtil.IsBoolPTrue(submitArgs.Command))
			submitArgs.Command = &isCommand

			err = applyTemplate(clientset, submitArgs, commandArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = submitArgs.setCommonRun(cmd, args, kubeClient, clientset)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if submitArgs.TtlAfterFinished != nil {
				ttlSeconds := int(math.Round(submitArgs.TtlAfterFinished.Seconds()))
				log.Debugf("Using time to live seconds %d", ttlSeconds)
				submitArgs.TTL = &ttlSeconds
			}

			if raUtil.IsBoolPTrue(submitArgs.IsJupyter) {
				submitArgs.UseJupyterDefaultValues()
			}

			err = submitRunaiJob(args, submitArgs, clientset, *runaijobClient)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			printJobInfoIfNeeded(submitArgs)
			if raUtil.IsBoolPTrue(submitArgs.IsJupyter) || (submitArgs.Interactive != nil && *submitArgs.Interactive && submitArgs.ServiceType == "portforward") {
				err = kubectl.WaitForReadyStatefulSet(submitArgs.Name, submitArgs.Namespace)

				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				if raUtil.IsBoolPTrue(submitArgs.IsJupyter) {
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

	fbg := flags.NewFlagsByGroups(command)

	submitArgs.addCommonFlags(fbg)
	submitArgs.addFlags(fbg)

	fbg.UpdateFlagsByGroupsToCmd()

	return command
}

func applyTemplate(clientset kubernetes.Interface, submitArgs *submitRunaiJobArgs, extraArgs []string) error {
	var err error
	configs := templates.NewTemplates(clientset)
	var templateToUse *templates.Template
	if templateName == "" {
		templateToUse, err = configs.GetDefaultTemplate()
	} else {
		templateToUse, err = configs.GetTemplate(templateName)
		if templateToUse == nil {
			return fmt.Errorf("could not find runai template %s. Please run '%s template list'", templateName, config.CLIName)
		}
	}

	if templateToUse != nil {
		err = applyTemplateToSubmitRunaijob(templateToUse.Values, submitArgs, extraArgs)
		if err != nil {
			return fmt.Errorf("could not apply template %s due to: %v", templateName, err)
		}
	}
	return nil
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
	IsJupyter        *bool
	IsPreemptible    *bool `yaml:"isPreemptible,omitempty"`
	IsRunaiJob       *bool `yaml:"isRunaiJob,omitempty"`
	IsOldJob         *bool
	TtlAfterFinished *time.Duration
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
	if len(sa.SpecCommand) == 0 && sa.ServiceType == "ingress" {
		sa.SpecCommand = []string{jupyterCommand}
		log.Infof("Using default jupyter notebook command for using ingress service \"%s\"", jupyterCommand)
	}
	if len(sa.SpecArgs) == 0 && sa.ServiceType == "ingress" {
		baseUrlArg := fmt.Sprintf(jupyterArgs, sa.Project, sa.Name)
		sa.SpecArgs = []string{baseUrlArg}
		log.Infof("Using default jupyter notebook command argument for using ingress service \"%s\"", baseUrlArg)
	}
}

// add flags to submit spark args
func (sa *submitRunaiJobArgs) addFlags(fbg flags.FlagsByGroups) {

	fs := fbg.GetOrAddFlagSet(JobLifecycleFlagGroup)
	fs.StringVarP(&(sa.ServiceType), "service-type", "s", "", "Specify service exposure for interactive jobs. Options are: portforward, loadbalancer, nodeport, ingress.")
	flags.AddBoolNullableFlag(fs, &(sa.IsJupyter), "jupyter", "", "Shortcut for running a jupyter notebook using a pre-created image and a default notebook configuration.")
	flags.AddBoolNullableFlag(fs, &(sa.Elastic), "elastic", "", "Mark the job as elastic.")
	flags.AddBoolNullableFlag(fs, &(sa.IsPreemptible), "preemptible", "", "Mark an interactive job as preemptible. Preemptible jobs can be scheduled above guaranteed quota but may be reclaimed at any time.")
	flags.AddIntNullableFlag(fs, &(sa.Completions), "completions", "The number of successful pods required for this job to be completed. Used for Hyperparameter optimization.")
	flags.AddIntNullableFlag(fs, &(sa.Parallelism), "parallelism", "The number of pods this job tries to run in parallel at any time.  Used for Hyperparameter optimization.")
	flags.AddIntNullableFlag(fs, &(sa.BackoffLimit), "backoffLimit", "The number of times the job will be retried before failing. Default 6.")
	flags.AddDurationNullableFlagP(fs, &(sa.TtlAfterFinished), "ttl-after-finish", "", "Define the duration, post job finish, after which the job is automatically deleted (e.g. 5s, 2m, 3h).")
	flags.AddBoolNullableFlag(fs, &(sa.IsOldJob), "old-job", "", "submit a job of resource k8s job")
	fs.MarkHidden("old-job")

	fs = fbg.GetOrAddFlagSet(NetworkFlagGroup)
	fs.StringArrayVar(&(sa.Ports), "port", []string{}, "Expose ports from the Job container.")

}

func submitRunaiJob(args []string, submitArgs *submitRunaiJobArgs, clientset kubernetes.Interface, runaiclientset runaiclientset.Clientset) error {
	err := verifyHPOFlags(submitArgs)
	if err != nil {
		return err
	}

	handleRunaiJobCRD(submitArgs, runaiclientset)
	submitArgs.Name, err = workflow.SubmitJob(submitArgs.Name, submitArgs.Namespace, submitArgs.generateSuffix, submitArgs, runaiChart, clientset, dryRun)
	if err != nil {
		return err
	}
	if !dryRun {
		fmt.Printf("The job '%s' has been submitted successfully\n", submitArgs.Name)
		fmt.Printf("You can run `%s get %s -p %s` to check the job status\n", config.CLIName, submitArgs.Name, submitArgs.Project)
	}

	return nil
}

// For backward compatibility - remove once all customers have runaijob crd
func handleRunaiJobCRD(submitArgs *submitRunaiJobArgs, runaiclientset runaiclientset.Clientset) {
	isRunaiJob := true
	submitArgs.IsRunaiJob = &isRunaiJob
	if submitArgs.IsOldJob != nil && *submitArgs.IsOldJob {
		*submitArgs.IsRunaiJob = false
		return
	}
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
	}
	return nil
}
