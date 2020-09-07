package commands

import (
	"fmt"
	"os"

	// "github.com/kubeflow/arena/cmd/arena/commands/flags"
	"github.com/kubeflow/arena/pkg/client"
	// "github.com/kubeflow/arena/pkg/util/kubectl"
	// log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	// "k8s.io/client-go/tools/remotecommand"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
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
			Attach(cmd, jobName)
		},
	}

	return cmd
}

// Attach to a running job name
func Attach(cmd *cobra.Command, interactive bool, name string,commandArgs []string, TTY bool,  podName string ) error {

	kubeClient, err := client.GetClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	podToExec, err := GetPodFromCmd(cmd, kubeClient, podName)

	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
	

	kubectl.Attach(podToExec.Name, podToExec.Namespace, commandArgs, interactive, TTY)


	// restClient, err := restclient.RESTClientFor(o.Config)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	req := restClient.Post().
	// 		Resource("pods").
	// 		Name(o.Pod.Name).
	// 		Namespace(o.Pod.Namespace).
	// 		SubResource("attach")
	// 	req.VersionedParams(&corev1.PodAttachOptions{
	// 		Container: containerToAttach.Name,
	// 		Stdin:     o.Stdin,
	// 		Stdout:    o.Out != nil,
	// 		Stderr:    !o.DisableStderr,
	// 		TTY:       raw,
	// 	}, scheme.ParameterCodec)
	
	
	// // method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) 
	// exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	// if err != nil {
	// 	return err
	// }
	// return exec.Stream(remotecommand.StreamOptions{
	// 	Stdin:             stdin,
	// 	Stdout:            stdout,
	// 	Stderr:            stderr,
	// 	Tty:               tty,
	// 	TerminalSizeQueue: terminalSizeQueue,
	// })
}