package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/kubeflow/arena/cmd/arena/commands/flags"
	raUtil "github.com/kubeflow/arena/cmd/arena/commands/util"
	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	"k8s.io/client-go/rest"

	// "k8s.io/cli-runtime/pkg/resource"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/remotecommand"
	kubeExec "k8s.io/kubectl/pkg/cmd/exec"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
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

			if err := Exec(cmd, name, []string{"/bin/bash"}, []string{}, DefaultAttachTimeout, true, true, podName, "bash"); err != nil {
				log.Error(err)
				os.Exit(1)
			}
		},
	}

	command.Flags().StringVar(&podName, "pod", "", "Specify a pod of a running job. To get a list of the pods of a specific job, run \"runai get <job-name>\" command")

	return command
}

func NewExecCommand() *cobra.Command {
	var interactive bool
	var TTY bool
	var podName string
	var fileNames []string

	var command = &cobra.Command{
		Use:   "exec JOB_NAME COMMAND [ARG ...]",
		Short: "Execute a command inside a running job.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {

			name = args[0]
			command := args[1:]

			if err := Exec(cmd, name, command, fileNames, DefaultAttachTimeout, interactive, TTY, podName, "exec"); err != nil {
				log.Error(err)
				os.Exit(1)
			}
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

func Exec(cmd *cobra.Command, jobName string, command, fileNames []string, timeout time.Duration, interactive bool, TTY bool, podName string, runaiCommandName string) (err error) {

	kubeClient, err := client.GetClient()
	if err != nil {
		return
	}

	foundPod, err := GetPodFromCmd(cmd, kubeClient, jobName, podName)

	if err != nil {
		return
	}

	pod, err := raUtil.WaitForPod(
		foundPod.Name,
		foundPod.Namespace,
		timeout,
		raUtil.NotReadyPodTimeoutMsg,
		raUtil.PodRunning,
	)

	if err != nil {
		return
	}

	return ExecByBin(pod, command[0], command[1:], interactive, TTY)

}

func ExecByLib(pod *v1.Pod, command, fileNames []string, stdin, tty bool) error {
	ioStream := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	restConfig, _ := matchVersionKubeConfigFlags.ToRESTConfig()

	o := &kubeExec.ExecOptions{
		StreamOptions: kubeExec.StreamOptions{
			Namespace: pod.Namespace,
			PodName:   pod.Name,
			IOStreams: ioStream,
			TTY:       tty,
			Stdin:     stdin,
		},

		Command:  command,
		Pod:      pod,
		Config:   restConfig,
		Executor: &kubeExec.DefaultRemoteExecutor{},
	}

	containerToAttach := &pod.Spec.Containers[0]
	t := o.SetupTTY()

	var sizeQueue remotecommand.TerminalSizeQueue
	if t.Raw {
		// this call spawns a goroutine to monitor/update the terminal size
		sizeQueue = t.MonitorSize(t.GetSize())

		// unset p.Err if it was previously set because both stdout and stderr go over p.Out when tty is
		// true
		o.ErrOut = nil
	}

	fn := func() error {
		restClient, err := rest.RESTClientFor(o.Config)
		if err != nil {
			return err
		}

		// TODO: consider abstracting into a client invocation or client helper
		req := restClient.Post().
			Resource("pods").
			Name(pod.Name).
			Namespace(pod.Namespace).
			SubResource("exec")
		req.VersionedParams(&corev1.PodExecOptions{
			Container: containerToAttach.Name,
			Command:   o.Command,
			Stdin:     o.Stdin,
			Stdout:    o.Out != nil,
			Stderr:    o.ErrOut != nil,
			TTY:       t.Raw,
		}, scheme.ParameterCodec)

		return o.Executor.Execute("POST", req.URL(), o.Config, o.In, o.Out, o.ErrOut, t.Raw, sizeQueue)
	}
	return t.Safe(fn)

}

func ExecByBin(pod *v1.Pod, command string, commandArgs []string, interactive, TTY bool) error {
	// NOTE: Getting a deprecation msg in some kubectl versions
	return kubectl.Exec(pod.Name, pod.Namespace, command, commandArgs, interactive, TTY)
}
