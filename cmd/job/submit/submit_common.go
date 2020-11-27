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

package submit

import (
	"fmt"
	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/pkg/submittionArgs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	dashArg                                   = "--"
	commandFlag                               = "command"
	oldCommandFlag                            = "old-command"
	NetworkFlagGroup      flags.FlagGroupName = "Network"
	JobLifecycleFlagGroup flags.FlagGroupName = "Job Lifecycle"
)

func mergeOldCommandAndArgsWithNew(argsLenAtDash int, positionalArgs, oldCommand, oldArgs []string, isCommand *bool) ([]string, *bool) {
	if argsLenAtDash == -1 {
		argsLenAtDash = len(positionalArgs)
	}

	argsAfterDash := positionalArgs[argsLenAtDash:]
	if len(argsAfterDash) != 0 {
		return argsAfterDash, isCommand
	}

	isAnyCommand := false
	if len(oldCommand) != 0 {
		isAnyCommand = true
	}
	return append(oldCommand, oldArgs...), &isAnyCommand
}

func convertOldCommandArgsFlags(cmd *cobra.Command, submitArgs *submittionArgs.SubmitArgs, args []string) []string {
	commandArgs, isCommand := mergeOldCommandAndArgsWithNew(cmd.ArgsLenAtDash(), args, submitArgs.SpecCommand, submitArgs.SpecArgs, submitArgs.Command)
	if isCommand != nil && *isCommand {
		submitArgs.SpecCommand = commandArgs
		submitArgs.SpecArgs = []string{}
	} else {
		submitArgs.SpecCommand = []string{}
		submitArgs.SpecArgs = commandArgs
	}
	submitArgs.Command = isCommand
	return commandArgs
}

func AlignArgsPreParsing(args []string) []string {
	if len(args) < 2 || (args[1] != submitCommand && args[1] != SubmitMpiCommand) {
		return args
	}

	dashIndex := -1
	for i, arg := range args {
		if arg == dashArg {
			dashIndex = i
		}
	}

	if dashIndex == -1 {
		for i, arg := range args {
			if arg == fmt.Sprintf("%s%s", dashArg, commandFlag) {
				log.Info(fmt.Sprintf("using %s%s as string flag has been deprecated. Please see usage information", dashArg, commandFlag))
				args[i] = fmt.Sprintf("%s%s", dashArg, oldCommandFlag)
			}
		}
	}
	return args
}
