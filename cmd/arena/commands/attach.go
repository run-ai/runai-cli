package commands

import (
	"fmt"
	"os"
	"io"
	netUrl "net/url"

	"k8s.io/apimachinery/pkg/runtime/schema"

	// "github.com/kubeflow/arena/cmd/arena/commands/flags"
	"github.com/kubeflow/arena/pkg/client"
	// "github.com/kubeflow/arena/pkg/util/kubectl"
	// log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/remotecommand"

	// "github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	kubeAttach "k8s.io/kubectl/pkg/cmd/attach"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	corev1 "k8s.io/api/core/v1"
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

			if err := Attach(cmd, jobName, true, true, ""); err != nil {
				log.Errorln(err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

// Attach to a running job name
func Attach(cmd *cobra.Command, jobName string, stdin, tty bool,  podName string ) (err error) {
	
	kubeClient, err := client.GetClient()
	if err != nil {
		return 
	}

	podToExec, err := GetPodFromCmd(cmd, kubeClient, jobName, podName)
	if err != nil {
		return 
	}

	ioStream := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr,}
	
	initIstioClient(kubeClient)
	
	o := kubeAttach.NewAttachOptions(ioStream)
	var sizeQueue remotecommand.TerminalSizeQueue
	t := o.SetupTTY()

	if podToExec == nil {
		return fmt.Errorf("Not found any matching pod")
	}

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)

	_ = cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// restClient, err := f.ToRESTConfig()

	restConfig, _ := initIstioClient(kubeClient)

	o.Pod = podToExec
	o.Namespace = podToExec.Namespace
	o.PodName = podToExec.Name
	o.TTY = tty
	o.Stdin = stdin

	o.Config = restConfig


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

	restClient, err := rest.RESTClientFor(o.Config)
	if err != nil {
		return err
	}
	req := restClient.Post().
		Resource("pods").
		Name(podToExec.Name).
		Namespace(podToExec.Namespace).
		SubResource("attach")
	req.VersionedParams(&corev1.PodAttachOptions{
		Container: containerToAttach.Name,
		Stdin:     stdin,
		Stdout:    o.Out != nil,
		Stderr:    !o.DisableStderr,
		TTY:       t.Raw,
	}, scheme.ParameterCodec)

	return DefaultAttach("POST", req.URL(), o.Config, o.In, o.Out, o.ErrOut, t.Raw, sizeQueue)

	// if err := t.Safe(o.AttachFunc(o, containerToAttach , t.Raw, sizeQueue)); err != nil {
	// 	return err
	// }

}

// DefaultAttach
func DefaultAttach(method string, url *netUrl.URL, config *rest.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
	fmt.Println("The url is", url)
	
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.Stream(remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               tty,
		TerminalSizeQueue: terminalSizeQueue,
	})
}


func initIstioClient(client *client.Client) (*rest.Config, error) {
	restConfig := client.GetRestConfig()

	apiGroupVersion := schema.GroupVersion{
		Version: "v1",
	}

	restConfig.GroupVersion = &apiGroupVersion
	

	restConfig.APIPath = "/apis"
	restConfig.ContentType = runtime.ContentTypeJSON

	types := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			metav1.AddToGroupVersion(scheme, apiGroupVersion)
			return nil
		})
	err := schemeBuilder.AddToScheme(types)
	if err!=nil {
		return nil, err
	}
	ns := serializer.CodecFactory{}
	ns.SupportedMediaTypes()
	restConfig.NegotiatedSerializer = ns
		//: serializer.NewCodecFactory(types),

	return restConfig, err
}