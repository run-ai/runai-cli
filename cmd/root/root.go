// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package root

import (
	"context"
	"github.com/run-ai/runai-cli/cmd/cluster"
	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/login"
	"github.com/run-ai/runai-cli/cmd/logout"
	"github.com/run-ai/runai-cli/cmd/resource"

	raCmd "github.com/run-ai/runai-cli/cmd"
	"github.com/run-ai/runai-cli/cmd/attach"
	"github.com/run-ai/runai-cli/cmd/exec"
	"github.com/run-ai/runai-cli/cmd/global"
	deleteJob "github.com/run-ai/runai-cli/cmd/job/delete"
	submitJob "github.com/run-ai/runai-cli/cmd/job/submit"
	"github.com/run-ai/runai-cli/cmd/logs"
	"github.com/run-ai/runai-cli/cmd/project"
	"github.com/run-ai/runai-cli/cmd/template"

	"github.com/run-ai/runai-cli/pkg/config"
	"github.com/run-ai/runai-cli/pkg/util"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// NewCommand returns a new instance of an Arena command
func NewCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   config.CLIName,
		Short: "runai is a command line interface to a RunAI cluster",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
		// Would be run before any child command
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(global.LogLevel)
		},
	}

	addKubectlFlagsToCmd(command)

	// enable logging
	command.PersistentFlags().StringVar(&global.LogLevel, "loglevel", "info", "Set the logging level. One of: debug|info|warn|error")

	command.AddCommand(submitJob.NewRunaiJobCommand())
	command.AddCommand(submitJob.NewRunaiSubmitMPIJobCommand())
	command.AddCommand(resource.NewListCommand())
	command.AddCommand(logs.NewLogsCommand())
	command.AddCommand(deleteJob.NewDeleteCommand())
	command.AddCommand(resource.GetCommand())
	command.AddCommand(resource.NewTopCommand())
	command.AddCommand(resource.NewDescribeCommand())
	command.AddCommand(resource.ConfigCommand())
	command.AddCommand(raCmd.NewVersionCmd())
	command.AddCommand(raCmd.NewUpdateCommand())
	command.AddCommand(exec.NewBashCommand())
	command.AddCommand(exec.NewExecCommand())
	command.AddCommand(attach.NewAttachCommand())
	command.AddCommand(template.NewTemplateCommand())
	command.AddCommand(project.NewProjectCommand())
	command.AddCommand(cluster.NewClusterCommand())
	command.AddCommand(login.NewLoginCommand())
	command.AddCommand(logout.NewLogoutCommand())

	return command
}

func addKubectlFlagsToCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(flags.ProjectFlag, "p", "", "Specify the project to which the command applies. By default, commands apply to the default project. To change the default project use ‘runai config project <project name>’.")
}

func createNamespace(client *kubernetes.Clientset, namespace string) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	return err
}

func getNamespace(client *kubernetes.Clientset, namespace string) (*v1.Namespace, error) {
	return client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
}

func ensureNamespace(client *kubernetes.Clientset, namespace string) error {
	_, err := getNamespace(client, namespace)
	if err != nil && errors.IsNotFound(err) {
		return createNamespace(client, namespace)
	}
	return err
}
