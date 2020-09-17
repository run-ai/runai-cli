package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/kubeflow/arena/cmd/arena/commands/flags"
	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
)

func NewBashCommand() *cobra.Command {
	var podName string
	var command = &cobra.Command{
		Use:   "bash JOB_NAME",
		Short: "Get a bash session inside a running job.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			name = args[0]

			execute(cmd, name, "/bin/bash", []string{}, true, true, podName, "bash")
		},
	}

	command.Flags().StringVar(&podName, "pod", "", "Specify a pod of a running job. To get a list of the pods of a specific job, run \"runai get <job-name>\" command")

	return command
}

func NewExecCommand() *cobra.Command {
	var interactive bool
	var TTY bool
	var podName string

	var command = &cobra.Command{
		Use:   "exec JOB_NAME COMMAND [ARG ...]",
		Short: "Execute a command inside a running job.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {

			name = args[0]
			command := args[1]
			commandArgs := args[2:]

			execute(cmd, name, command, commandArgs, interactive, TTY, podName, "exec")
		},
	}

	command.Flags().StringVar(&podName, "pod", "", "Specify a pod of a running job. To get a list of the pods of a specific job, run \"runai get <job-name>\" command")
	command.Flags().BoolVarP(&interactive, "stdin", "i", false, "Pass stdin to the container")
	command.Flags().BoolVarP(&TTY, "tty", "t", false, "Stdin is a TTY")

	return command
}

// GetPodFromCmd extract and searche for namespace, job and podName
func GetPodFromCmd(cmd *cobra.Command, kubeClient *client.Client, jobName, podName string) (pod *v1.Pod, err error) {

	namespace, err := flags.GetNamespaceToUseFromProjectFlag(cmd, kubeClient)

	if err != nil {
		return
	}

	job, err := searchTrainingJob(kubeClient, jobName, "", namespace)
	if err != nil {
		return
	} else if job == nil {
		err = fmt.Errorf("The job: '%s' is not found", jobName)
		return
	}

	if len(podName) == 0 {
		pod = job.ChiefPod()
	} else {
		pods := job.AllPods()
		for _, p := range pods {
			if podName == p.Name {
				pod = &p
				break
			}
		}
	}

	if pod == nil {
		err = fmt.Errorf("Failed to find pod: '%s' of job: '%s'", podName, job.Name())
	}

	return 
}

const (
	NotReadyPodTimeoutMsg = "Timeout .. Please wait until the job is running and try again"
)

// WaitForPod waiting to the pod phase to become running
func WaitForPod(getPod func() (*v1.Pod, error), timeout time.Duration, timeoutMsg string, exitCondition func(*v1.Pod, int) (bool, error) ) ( pod *v1.Pod, err error)  {
	shouldStopAt := time.Now().Add( timeout)

	for i, exit := 0, false;; i++ {
		pod, err = getPod()
		if err != nil {
			return 
		}

		exit, err = exitCondition(pod, i)
		if err != nil || exit {
			return 
		}

		if shouldStopAt.Before( time.Now()) {
			return nil, fmt.Errorf(timeoutMsg)
		}
		time.Sleep(time.Second)	
	}
}


func PodRunning(pod *v1.Pod, i int) (exit bool, err error) {
	phase := pod.Status.Phase

	switch phase {
	case v1.PodPending:
		break
	case v1.PodRunning:
		conditions := pod.Status.Conditions
		if conditions == nil {
			return false, nil
		}
		for i := range conditions {
			if conditions[i].Type == corev1.PodReady &&
				conditions[i].Status == corev1.ConditionTrue {
					exit = true 
			}
		}
		
	default:
		err = fmt.Errorf("Can't connect to the pod: %s in phase: %s",pod.Name, phase)
	}

	if exit {
		if i > 0 {
			fmt.Print("\n")
		}
	} else if i == 0 {
		fmt.Print("Waiting...")
	} else {
		fmt.Print(".")
	}
	return
}


func execute(cmd *cobra.Command, jobName string, command string, commandArgs []string, interactive bool, TTY bool, podName string, runaiCommandName string) {

	kubeClient, err := client.GetClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	podToExec, err := WaitForPod(
		func() (*v1.Pod, error) { return GetPodFromCmd(cmd, kubeClient, jobName, podName)}, 
		time.Second * 10,
		NotReadyPodTimeoutMsg,
		PodRunning,
	)

	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}

	kubectl.Exec(podToExec.Name, podToExec.Namespace, command, commandArgs, interactive, TTY)
}
