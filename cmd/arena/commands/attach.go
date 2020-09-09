package commands

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime/schema"

	// "github.com/kubeflow/arena/cmd/arena/commands/flags"
	"github.com/kubeflow/arena/pkg/client"
	// "github.com/kubeflow/arena/pkg/util/kubectl"
	// log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	o.Pod = podToExec
	o.Namespace = podToExec.Namespace
	o.PodName = podToExec.Name
	o.TTY = tty
	o.Stdin = stdin
	restClient, err := f.ToRESTConfig()

	o.Config = restClient


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


func initIstioClient(client *client.Client) (*rest.RESTClient, error) {
	restConfig := client.GetRestConfig()

	istioAPIGroupVersion := schema.GroupVersion{
		Group:   "networking.istio.io",
		Version: "v1alpha3",
	}
	//istioAPIGroupVersion := schema.GroupVersion{
	//	Group:   "config.istio.io",
	//	Version: "v1alpha2",
	//}



	restConfig.GroupVersion = &istioAPIGroupVersion

	restConfig.APIPath = "/apis"
	restConfig.ContentType = runtime.ContentTypeJSON

	types := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			metav1.AddToGroupVersion(scheme, istioAPIGroupVersion)
			return nil
		})
	err := schemeBuilder.AddToScheme(types)
	if err!=nil {
		return nil, err
	}

	// restConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(types)}

	return rest.RESTClientFor(restConfig)
}