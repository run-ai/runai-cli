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

// +build ignore

package serve

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"bytes"
	"io/ioutil"

	"github.com/run-ai/runai-cli/cmd/global"
	"github.com/run-ai/runai-cli/pkg/util"
	"github.com/run-ai/runai-cli/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	tfservingChart        = util.GetChartsFolder() + "/tfserving"
	defaultTfServingImage = "tensorflow/serving:latest"
)

func NewServingTensorFlowCommand() *cobra.Command {
	var (
		serveTensorFlowArgs ServeTensorFlowArgs
	)

	var command = &cobra.Command{
		Use:     "tensorflow",
		Short:   "Submit tensorflow serving job to deploy and serve machine learning models.",
		Aliases: []string{"tf"},
		PreRun: commandUtil.WrapRoleAssertion(auth.AssertExecutorRole),
		Run: func(cmd *cobra.Command, args []string) {
			/*if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}*/

			if serveTensorFlowArgs.GPUMemory != 0 && serveTensorFlowArgs.GPUCount != 0 {
				fmt.Println("gpucount and gpumemory should not be used at the same time.You can only choose one mode")
				os.Exit(1)
			}
			util.SetLogLevel(global.LogLevel)
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			err = serveTensorFlow(args, &serveTensorFlowArgs, client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	serveTensorFlowArgs.addServeCommonFlags(command)

	// TFServingJob
	// add grpc port and rest api port
	command.Flags().StringVar(&serveTensorFlowArgs.Image, "image", defaultTfServingImage, "the docker image name of serve job, and the default image is "+defaultTfServingImage)
	command.Flags().IntVar(&serveTensorFlowArgs.Port, "port", 8500, "the port of tensorflow gRPC listening port")
	command.Flags().IntVar(&serveTensorFlowArgs.RestfulPort, "restfulPort", 8501, "the port of tensorflow RESTful listening port")

	command.Flags().StringVar(&serveTensorFlowArgs.ModelName, "modelName", "", "the model name for serving")
	command.Flags().MarkDeprecated("modelName", "please use --model-name instead")
	command.Flags().StringVar(&serveTensorFlowArgs.ModelName, "model-name", "", "the model name for serving")

	command.Flags().StringVar(&serveTensorFlowArgs.ModelPath, "modelPath", "", "the model path for serving in the container")
	command.Flags().MarkDeprecated("modelPath", "please use --model-path instead")
	command.Flags().StringVar(&serveTensorFlowArgs.ModelPath, "model-path", "", "the model path for serving in the container")

	command.Flags().StringVar(&serveTensorFlowArgs.ModelConfigFile, "modelConfigFile", "", "Corresponding with --model_config_file in tensorflow serving")
	command.Flags().StringVar(&serveTensorFlowArgs.VersionPolicy, "versionPolicy", "", "support latest, latest:N, specific:N, all")
	command.Flags().StringVar(&serveTensorFlowArgs.Command, "command", "", "the command will inject to container's command.")
	return command
}

type ServeTensorFlowArgs struct {
	VersionPolicy          string `yaml:"versionPolicy"`   // --versionPolicy
	ModelConfigFile        string `yaml:"modelConfigFile"` // --modelConfigFile
	ModelConfigFileContent string `yaml:"modelConfigFileContent"`
	Image                  string `yaml:"image"`     // --image
	ModelName              string `yaml:"modelName"` // --modelName
	ModelPath              string `yaml:"modelPath"` // --modelPath

	ServeArgs `yaml:",inline"`

	// ModelServiceExists bool `yaml:"modelServiceExists"` // --modelServiceExists
}

func (serveTensorFlowArgs ServeTensorFlowArgs) validateModelName() error {
	if serveTensorFlowArgs.ModelName == "" {
		return fmt.Errorf("--modelName cannot be blank")
	}

	var reg *regexp.Regexp
	reg = regexp.MustCompile(regexp4serviceName)
	matched := reg.MatchString(serveTensorFlowArgs.ModelName)
	if !matched {
		return fmt.Errorf("--modelName should be numbers, letters, dashes, and underscores ONLY")
	}

	return nil
}

func (serveTensorFlowArgs *ServeTensorFlowArgs) preprocess(client *kubernetes.Clientset, args []string) (err error) {
	//serveTensorFlowArgs.Command = strings.Join(args, " ")
	log.Debugf("command: %s", serveTensorFlowArgs.Command)

	if serveTensorFlowArgs.ModelConfigFile == "" {
		// need to validate modelName, modelPath and versionPolicy if not specify modelConfigFile
		// 1. validate modelName
		err := serveTensorFlowArgs.validateModelName()
		if err != nil {
			return err
		}
		//2. validate modelPath
		if serveTensorFlowArgs.ModelPath == "" {
			return fmt.Errorf("modelPath should be specified if no modelConfigFile is specified")
		}

		//3. validate versionPolicy
		err = serveTensorFlowArgs.validateVersionPolicy()
		if err != nil {
			return err
		}
		//populate content according to CLI parameters
		serveTensorFlowArgs.ModelConfigFileContent = generateModelConfigFileContent(*serveTensorFlowArgs)

	} else {
		//populate content from modelConfigFile
		if serveTensorFlowArgs.ModelName != "" {
			return fmt.Errorf("modelConfigFile=%s is specified, so --modelName cannot be used", serveTensorFlowArgs.ModelConfigFile)
		}
		if serveTensorFlowArgs.ModelPath != "" {
			return fmt.Errorf("modelConfigFile=%s is specified, so --modelPath cannot be used", serveTensorFlowArgs.ModelConfigFile)
		}

		modelConfigFileContentBytes, err := ioutil.ReadFile(serveTensorFlowArgs.ModelConfigFile)
		if err != nil {
			return fmt.Errorf("cannot read the modelConfigFile[%s]: %s", serveTensorFlowArgs.ModelConfigFile, err)
		}
		modelConfigString := string(modelConfigFileContentBytes)
		log.Debugf("The content of modelConfigFile[%s] is: %s", serveTensorFlowArgs.ModelConfigFile, modelConfigString)
		serveTensorFlowArgs.ModelConfigFileContent = modelConfigString
	}
	// validate models data
	if len(dataset) > 0 {
		err := ParseMountPath(dataset)
		if err != nil {
			return fmt.Errorf("--data has wrong value: %s", err)
		}
		serveTensorFlowArgs.ModelDirs = transformSliceToMap(dataset, ":")
	}

	log.Debugf("models:%s", serveTensorFlowArgs.ModelDirs)

	//validate Istio enablement
	err = serveTensorFlowArgs.ServeArgs.validateIstioEnablement()
	if err != nil {
		return err
	}

	// populate environment variables
	if len(envs) > 0 {
		serveTensorFlowArgs.Envs = transformSliceToMap(envs, "=")
	}
	if serveTensorFlowArgs.ServingVersion == "" {
		t := time.Now()
		serveTensorFlowArgs.ServingVersion = fmt.Sprint(t.Format("200601021504"))
	}
	modelServiceExists, err := checkServiceExists(client, namespace, serveTensorFlowArgs.ServingName)
	serveTensorFlowArgs.ModelServiceExists = modelServiceExists

	return nil
}

func checkServiceExists(client *kubernetes.Clientset, namespace string, name string) (found bool, err error) {
	service, err := client.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("service %s in namespace %s due to %v is not found.", name, namespace, err)
			err = nil
		}
		return found, err
	}

	if service != nil {
		found = true
	}

	return found, err
}

func (serveTensorFlowArgs *ServeTensorFlowArgs) validateVersionPolicy() error {
	// validate version policy
	if serveTensorFlowArgs.VersionPolicy == "" {
		serveTensorFlowArgs.VersionPolicy = "latest"
	}
	versionPolicyName := strings.Split(serveTensorFlowArgs.VersionPolicy, ":")
	switch versionPolicyName[0] {
	case "latest", "specific", "all":
		log.Debug("Support TensorFlow Serving Version Policy: latest, specific, all.")
		//serveTensorFlowArgs.ServeArgs.ModelVersion = strings.Replace(serveTensorFlowArgs.VersionPolicy, ":", "-", -1)
	default:
		return fmt.Errorf("UnSupport TensorFlow Serving Version Policy: %s", versionPolicyName[0])
	}

	return nil
}

func serveTensorFlow(args []string, serveTensorFlowArgs *ServeTensorFlowArgs, client *kubernetes.Clientset) (err error) {
	err = serveTensorFlowArgs.preprocess(client, args)
	if err != nil {
		return err
	}

	name = serveTensorFlowArgs.ServingName
	if serveTensorFlowArgs.ServingVersion != "" {
		name += "-" + serveTensorFlowArgs.ServingVersion
	}
	return workflow.SubmitJob(name, "tf-serving", namespace, serveTensorFlowArgs, "", tfservingChart, clientset)
}

func generateModelConfigFileContent(serveTensorFlowArgs ServeTensorFlowArgs) string {
	modelName := serveTensorFlowArgs.ModelName
	versionPolicy := serveTensorFlowArgs.VersionPolicy
	mountPath := serveTensorFlowArgs.ModelPath
	versionPolicyName := strings.Split(versionPolicy, ":")

	var buffer bytes.Buffer
	buffer.WriteString("model_config_list: { config: { name: ")
	buffer.WriteString("\"" + modelName + "\" base_path: \"")
	buffer.WriteString(mountPath + "\" model_platform: \"")
	buffer.WriteString("tensorflow\" model_version_policy: { ")
	switch versionPolicyName[0] {
	case "all":
		buffer.WriteString(versionPolicyName[0] + ": {} } } }")
	case "specific":
		if len(versionPolicyName) > 1 {
			buffer.WriteString(versionPolicyName[0] + ": { " + "versions: " + versionPolicyName[1] + " } } } }")
		} else {
			log.Errorf("[specific] version policy scheme should be specific:N")
		}
	case "latest":
		if len(versionPolicyName) > 1 {
			buffer.WriteString(versionPolicyName[0] + ": { " + "num_versions: " + versionPolicyName[1] + " } } } }")
		} else {
			buffer.WriteString(versionPolicyName[0] + ": { " + "num_versions: 1 } } } }")
		}
	default:
		log.Errorf("UnSupport TensorFlow Serving Version Policy: %s", versionPolicyName[0])
		buffer.Reset()
	}

	result := buffer.String()
	log.Debugf("generateModelConfigFileContent: \n%s", result)

	return result
}
