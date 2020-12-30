package attach

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"os"
	"time"

	"github.com/run-ai/runai-cli/cmd/exec"
	raUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/remotecommand"
	kubeAttach "k8s.io/kubectl/pkg/cmd/attach"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// AttachOptions contains the option for attach command
type AttachOptions struct {
	NoTTY   bool
	NoStdIn bool
	PodName string
}

// DefaultAttachTimeout ..
const DefaultAttachTimeout = time.Second * 30

// NewAttachCommand creating a new attach command
func NewAttachCommand() *cobra.Command {
	options := AttachOptions{}

	cmd := &cobra.Command{
		Use:    "attach JOB_NAME",
		Short:  "Attach standard input, output, and error streams to a running job session.",
		PreRun: commandUtil.NamespacedRoleAssertion(assertion.AssertExecutorRole),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			jobName := args[0]

			if err := Attach(cmd, jobName, !options.NoStdIn, !options.NoTTY, options.PodName, DefaultAttachTimeout); err != nil {
				log.Errorln(err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVarP(&(options.NoStdIn), "no-stdin", "", false, "Not pass stdin to the container")
	cmd.Flags().BoolVarP(&(options.NoTTY), "no-tty", "", false, "Not allocated a tty")
	cmd.Flags().StringVar(&(options.PodName), "pod", "", "Specify a pod of a running job. To get a list of the pods of a specific job, run \"runai get <job-name>\" command")

	return cmd
}

// Attach attach to a running job name
func Attach(cmd *cobra.Command, jobName string, stdin, tty bool, podName string, timeout time.Duration) (err error) {
	kubeClient, err := client.GetClient()
	if err != nil {
		return
	}

	foundPod, err := exec.GetPodFromCmd(cmd, kubeClient, jobName, podName)

	if err != nil {
		return
	}

	podToExec, err := raUtil.WaitForPod(
		foundPod.Name,
		foundPod.Namespace,
		raUtil.NotReadyPodWaitingMsg,
		timeout,
		raUtil.NotReadyPodTimeoutMsg,
		raUtil.PodRunning,
	)

	if err != nil {
		return
	} else if podToExec == nil {
		return fmt.Errorf("Not found any matching pod")
	}

	if podName == "" {
		// notify the user which pod name he will to attach
		fmt.Println("Connecting to pod", podToExec.Name)
	}

	return attachByKubectlLib(podToExec, stdin, tty)
}

// attachByKubeCtlBin attach to a running job name
func attachByKubeCtlBin(pod *v1.Pod, stdin, tty bool) (err error) {
	return kubectl.Attach(pod.Name, pod.Namespace, stdin, tty)
}

// attachByKubectlLib Attach to a running job name
func attachByKubectlLib(pod *v1.Pod, stdin, tty bool) (err error) {

	var sizeQueue remotecommand.TerminalSizeQueue
	ioStream := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	restConfig, err := matchVersionKubeConfigFlags.ToRESTConfig()

	o := kubeAttach.NewAttachOptions(ioStream)
	o.Pod = pod
	o.Namespace = pod.Namespace
	o.PodName = pod.Name
	o.TTY = tty
	o.Stdin = stdin
	o.Config = restConfig

	t := o.SetupTTY()
	containerToAttach := &pod.Spec.Containers[0]

	if o.TTY && !containerToAttach.TTY {
		return fmt.Errorf("Unable to use a TTY - container %s did not allocate one", containerToAttach.Name)

	} else if !o.TTY && containerToAttach.TTY {
		// the container was launched with a TTY, so we have to force a TTY here, otherwise you'll get
		// an error "Unrecognized input header"
		o.TTY = true
	}

	if t.Raw {
		if size := t.GetSize(); size != nil {
			// fake resizing +1 and then back to normal so that attach-detach-reattach will result in the
			// screen being redrawn
			sizePlusOne := *size
			sizePlusOne.Width++
			sizePlusOne.Height++
			// this call spawns a goroutine to monitor/update the terminal size
			sizeQueue = t.MonitorSize(&sizePlusOne, size)
		}

		o.DisableStderr = true
	}

	if !o.Quiet {
		fmt.Fprintln(o.ErrOut, "If you don't see a command prompt, try pressing enter.")
	}

	return t.Safe(o.AttachFunc(o, containerToAttach, t.Raw, sizeQueue))
}
