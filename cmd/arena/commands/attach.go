package commands

import (
	"fmt"
	"os"

	// "github.com/kubeflow/arena/cmd/arena/commands/flags"
	"github.com/kubeflow/arena/pkg/client"
	// "github.com/kubeflow/arena/pkg/util/kubectl"
	// log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	// "github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	kubeAttach "k8s.io/kubectl/pkg/cmd/attach"
	// restclient "k8s.io/client-go/rest"
	// v1 "k8s.io/api/core/v1"
)

// AttachOptions contains the option for attach command
type AttachOptions struct {

}

// NewAttachCommand creating a new attach command
func NewAttachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach JOB_NAME",
		Short: "Attach a running job session.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			jobName := args[0]

			fmt.Println(`hi from attach command`, args)
			Attach(cmd, jobName, true, true, "")
		},
	}

	return cmd
}

// Attach to a running job name
func Attach(cmd *cobra.Command, jobName string, stdin, tty bool,  podName string ) error {
	
	kubeClient, err := client.GetClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	podToExec, err := GetPodFromCmd(cmd, kubeClient, jobName, podName)

	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}

	ioStream := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr,}

	o := kubeAttach.NewAttachOptions(ioStream)

	var sizeQueue remotecommand.TerminalSizeQueue
	t := o.SetupTTY()

	o.Pod = podToExec
	o.Namespace = podToExec.Namespace
	o.PodName = podToExec.Name
	o.TTY = tty
	o.Stdin = stdin

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
	containerToAttach :=&podToExec.Spec.Containers[0]

	if !o.Quiet {
		fmt.Fprintln(o.ErrOut, "If you don't see a command prompt, try pressing enter.")
	}

	if err := t.Safe(o.AttachFunc(o, containerToAttach , t.Raw, sizeQueue)); err != nil {
		return err
	}

	return nil
}