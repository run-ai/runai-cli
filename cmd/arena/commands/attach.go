package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/kubeflow/arena/pkg/util/kubectl"
	"github.com/kubeflow/arena/pkg/client"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/remotecommand"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	kubeAttach "k8s.io/kubectl/pkg/cmd/attach"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// AttachOptions contains the option for attach command
type AttachOptions struct {
	NoTTY bool
	NoStdIn bool
	PodName string
}

// DefaultAttachTimeout .. 
const DefaultAttachTimeout = time.Second * 30

// NewAttachCommand creating a new attach command
func NewAttachCommand() *cobra.Command {
	options := AttachOptions{}

	cmd := &cobra.Command{
		Use:   "attach JOB_NAME",
		Short: "Attach standard input, output, and error streams to a running job session.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			jobName := args[0]

			if err := Attach(cmd, jobName, !options.NoStdIn, !options.NoTTY, options.PodName, DefaultAttachTimeout ); err != nil {
				log.Errorln(err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVarP(&(options.NoStdIn), "no-stdin", "", false, "Not pass stdin to the container")
	cmd.Flags().BoolVarP(&(options.NoTTY), "no-tty", "", false, "Not allocated a tty")
	cmd.Flags().StringVar(&(options.PodName), "pod", "", "Which pod to connect, by default connect to the chief pod")

	return cmd
}

// Attach attach to a running job name
func Attach(cmd *cobra.Command, jobName string, stdin, tty bool,  podName string, timeout time.Duration  ) (err error) {
	kubeClient, err := client.GetClient()
	if err != nil {
		return 
	}

	podToExec, err := WaitForPod(
		func() (*v1.Pod, error) { return GetPodFromCmd(cmd, kubeClient, jobName, podName) },
		timeout,
	)

	if err != nil {
		return 
	} else if podToExec == nil {
		return fmt.Errorf("Not found any matching pod")
	}

	if podName == "" {
		// notify the user which pod name he will to attach
		fmt.Println("Trying to connect to a pod called:", podToExec.Name)
	}

	return attachByKubectlLib(podToExec, stdin, tty)
}

// attachByKubeCtlBin attach to a running job name 
func attachByKubeCtlBin(pod *v1.Pod, stdin, tty bool, timeout time.Duration  ) (err error) { 
	return kubectl.Attach(pod.Name, pod.Namespace, stdin, tty)
}

// attachByKubectlLib Attach to a running job name
func attachByKubectlLib(pod *v1.Pod, stdin, tty bool ) (err error) {
	
	var sizeQueue remotecommand.TerminalSizeQueue
	ioStream := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr,}
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
	containerToAttach :=&pod.Spec.Containers[0]

	if !o.Quiet {
		fmt.Fprintln(o.ErrOut, "If you don't see a command prompt, try pressing enter.")
	}

	return t.Safe(o.AttachFunc(o, containerToAttach, t.Raw, sizeQueue));
}

