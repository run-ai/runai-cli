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

package resource

import (
	"github.com/spf13/cobra"
	"github.com/run-ai/runai-cli/cmd/node"
	"github.com/run-ai/runai-cli/cmd/job"

	// podv1 "k8s.io/api/core/v1"
)


func NewTopCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "top",
		Short: "Display top information about resources.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	// create subcommands
	command.AddCommand(node.NewTopNodeCommand())
	command.AddCommand(job.NewTopJobCommand())

	return command
}
