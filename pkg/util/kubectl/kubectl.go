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

package kubectl

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	util "github.com/kubeflow/arena/pkg/util"

	"github.com/kubeflow/arena/pkg/types"
	log "github.com/sirupsen/logrus"
)

var kubectlCmd = []string{"kubectl"}

func SupportOldDryRun() (bool, error) {
	args := []string{"version", "--client", "--short"}
	out, err := kubectl(args)
	if err != nil {
		return false, err
	}

	output := string(out)

	re, err := regexp.Compile("v(.*?)\\.(.*)\\..*")
	if err != nil {
		return false, err
	}

	errTemplate := fmt.Errorf("Could not find kubectl command version")

	res := re.FindStringSubmatch(output)
	if len(res) < 3 {
		return false, errTemplate
	}

	majorVersion, err := strconv.Atoi(res[1])
	if err != nil {
		return false, errTemplate
	}

	minorVersion, err := strconv.Atoi(res[2])
	if err != nil {
		return false, errTemplate
	}

	if majorVersion < 1 || (majorVersion == 1 && minorVersion < 18) {
		return true, nil
	} else {
		return false, nil
	}
}

/**
* dry-run creating kubernetes App Info for delete in future
* Exec /usr/local/bin/kubectl, [create --dry-run -f /tmp/values313606961 --namespace default]
**/

func SaveAppInfo(fileName, namespace string) (configFileName string, err error) {
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		return "", err
	}

	supportOldDryRun, err := SupportOldDryRun()
	if err != nil {
		return "", err
	}

	var args []string

	if supportOldDryRun {
		args = []string{"create", "--dry-run", "--namespace", namespace, "-f", fileName}
	} else {
		args = []string{"create", "--dry-run=client", "--namespace", namespace, "-f", fileName}
	}

	out, err := kubectl(args)
	output := string(out)
	result := []string{}

	// fmt.Printf("%s\n", string(out))
	if err != nil {
		log.Errorf("Failed to execute %s, %v with %v", "kubectl", args, err)
		log.Errorf("The output is %s\n", output)
		return "", err
	}

	// 1. generate the config file
	configFile, err := ioutil.TempFile("", "config")
	if err != nil {
		log.Errorf("Failed to create tmp file %v due to %v", configFile.Name(), err)
		return "", err
	}

	configFileName = configFile.Name()
	log.Debugf("Save the config file %s", configFileName)

	// 2. save app types to config file
	lines := strings.Split(output, "\n")
	log.Debugf("dry run result: %v", lines)

	for _, line := range lines {
		line := strings.TrimSpace(line)
		cols := strings.Fields(line)
		log.Debugf("cols: %s, %d", cols, len(cols))
		if len(cols) == 0 {
			continue
		}
		result = append(result, cols[0])
	}

	data := []byte(strings.Join(result, "\n"))
	defer configFile.Close()
	_, err = configFile.Write(data)
	if err != nil {
		log.Errorf("Failed to write %v to %s due to %v", data, configFileName, err)
		return configFileName, err
	}

	return configFileName, nil
}

/**
* Delete kubernetes config to uninstall app
* Exec /usr/local/bin/kubectl, [delete -f /tmp/values313606961 --namespace default]
**/
func UninstallApps(fileName, namespace string) (err error) {
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		return err
	}

	args := []string{"delete", "--namespace", namespace, "-f", fileName}
	out, err := kubectl(args)

	fmt.Printf("%s\n", string(out))
	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return err
}

/**
* Delete kubernetes config to uninstall app
* Exec /usr/local/bin/kubectl, [delete -f /tmp/values313606961 --namespace default]
**/
func UninstallAppsWithAppInfoFile(appInfoFile, namespace string) (output string, err error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return "", err
	}

	if _, err = os.Stat(appInfoFile); err != nil {
		return "", err
	}

	args := []string{"cat", appInfoFile, "|", "xargs",
		binary, "delete"}

	args = util.AddNamespaceToArgs(args, namespace)

	log.Debugf("Exec bash -c %v", args)

	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	env := os.Environ()
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}
	out, err := cmd.CombinedOutput()
	log.Debugf("%s", string(out))

	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "bash -c", args, err)
	}

	return string(out), err
}

/**
* Apply kubernetes config to install app
* Exec /usr/local/bin/kubectl, [apply -f /tmp/values313606961 --namespace default]
**/
func InstallApps(fileName, namespace string) (output string, err error) {
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		return output, err
	}

	args := []string{"apply", "--namespace", namespace, "-f", fileName}
	out, err := kubectl(args)

	log.Debugf("%s", string(out))
	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return string(out), err
}

/**
* This name should be <job-type>-<job-name>
* create configMap by using name, namespace and configFile
**/
func CreateAppConfigmap(jobName, namespace, configFileName, envValuesFile, appInfoFileName, chartName, chartVersion string) (err error) {
	if _, err = os.Stat(configFileName); os.IsNotExist(err) {
		return err
	}

	if _, err = os.Stat(appInfoFileName); os.IsNotExist(err) {
		return err
	}

	args := []string{"create", "configmap", jobName,
		"--namespace", namespace,
		fmt.Sprintf("--from-file=%s=%s", "values", configFileName),
		fmt.Sprintf("--from-file=%s=%s", "app", appInfoFileName),
		fmt.Sprintf("--from-literal=%s=%s", chartName, chartVersion)}
	// "--overrides='{\"metadata\":{\"label\":\"createdBy\": \"arena\"}}'"}
	if envValuesFile != "" {
		args = append(args, fmt.Sprintf("--from-file=%s=%s", "env-values", envValuesFile))
	}

	out, err := kubectl(args)

	log.Debugf("%s", string(out))
	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return err
}

func LabelAppConfigmap(jobName, namespace, label string) (err error) {
	args := []string{"label", "configmap", jobName,
		"--namespace", namespace,
		label}
	// "--overrides='{\"metadata\":{\"label\":\"createdBy\": \"arena\"}}'"}
	out, err := kubectl(args)

	log.Debugf("%s", string(out))
	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return err
}

func CheckIfAppInfofileContentsExists(appFileName string, namespace string) (bool, error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return false, err
	}

	if _, err = os.Stat(appFileName); err != nil {
		return false, err
	}

	resourcesString, err := ioutil.ReadFile(appFileName)
	if err != nil {
		return false, err
	}

	resources := strings.Split(string(resourcesString), " ")

	args := []string{"get"}
	args = append(args, resources...)

	args = util.AddNamespaceToArgs(args, namespace)

	log.Debugf("kubectl %v", args)

	cmd := exec.Command(binary, args...)
	env := os.Environ()
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}

	err = cmd.Run()
	// Error should be returned if one of the resources does not exists on kubernetes
	if err != nil {
		return false, nil
	}

	return true, nil
}

/**
*
* delete configMap by using name, namespace
**/
func DeleteAppConfigMap(name, namespace string) (err error) {
	args := []string{"delete", "configmap", name, "--namespace", namespace}
	out, err := kubectl(args)

	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
		log.Debugf("%s", string(out))
	} else {
		log.Debugf("%s", string(out))
	}

	return err
}

/**
*
* get configMap by using name, namespace
**/
func CheckAppConfigMap(name, namespace string) (found bool) {
	args := []string{"get", "configmap", name}
	args = util.AddNamespaceToArgs(args, namespace)
	out, err := kubectl(args)

	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
		log.Debugf("%s", string(out))
	} else {
		log.Debugf("%s", string(out))
		found = true
	}

	return found
}

/**
*
* save the key of configMap into a file
**/
func SaveAppConfigMapToFile(name, key, namespace string) (fileName string, err error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return "", err
	}

	file, err := ioutil.TempFile(os.TempDir(), name)
	if err != nil {
		log.Errorf("Failed to create tmp file %v due to %v", file.Name(), err)
		return fileName, err
	}
	fileName = file.Name()

	args := []string{binary, "get", "configmap", name, fmt.Sprintf("-o=jsonpath='{.data.%s}'", key)}
	args = util.AddNamespaceToArgs(args, namespace)
	args = append(args, ">", fileName)

	log.Debugf("Exec bash -c %s", strings.Join(args, " "))

	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	env := os.Environ()
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}
	out, err := cmd.Output()
	fmt.Printf("%s", string(out))

	if err != nil {
		return fileName, fmt.Errorf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}
	return fileName, err
}

func WaitForReadyStatefulSet(name string, namespace string) error {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return err
	}

	env := os.Environ()
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}

	log.Infof("Waiting for job to start")
	args := []string{"-c", fmt.Sprintf("while [ $(%s get statefulset %s -n %s -o custom-columns=READY:.status.readyReplicas --no-headers ) != \"1\" ]; do echo \"Waiting for job to start\" && sleep 5; done", binary, name, namespace)}
	cmd := exec.Command("bash", args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()

	if err != nil {
		return err
	}

	log.Infof("Job started")
	return nil
}

func Attach(podName string, namespace string, commandArgs []string, interactive bool, TTY bool) error {
	args := []string{"attach", podName, fmt.Sprintf("-i=%t", interactive), fmt.Sprintf("-t=%t", TTY), "-n", namespace}
	args = append(args, commandArgs...)
	return kubectlAttched(args)
}

func Exec(podName string, namespace string, command string, commandArgs []string, interactive bool, TTY bool) error {
	args := []string{"exec", podName, fmt.Sprintf("-i=%t", interactive), fmt.Sprintf("-t=%t", TTY), "-n", namespace, command}
	args = append(args, commandArgs...)
	return kubectlAttched(args)
}

func Logs(podName string, namespace string) (string, error) {
	args := []string{"logs", podName, "-n", namespace}
	return kubectl(args)
}

func PortForward(ports []string, serviceName string, namespace string) error {
	args := []string{"port-forward", fmt.Sprintf("service/%s", serviceName), "--pod-running-timeout=1m0s", "-n", namespace}
	args = append(args, ports...)
	return kubectlAttched(args)
}

func kubectlAttched(args []string) error {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return err
	}

	// 1. prepare the arguments
	// args := []string{"create", "configmap", name, "--namespace", namespace, fmt.Sprintf("--from-file=%s=%s", name, configFileName)}
	log.Debugf("Exec %s, %v", binary, args)

	env := os.Environ()
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}

	// return syscall.Exec(cmd, args, env)
	// 2. execute the command
	cmd := exec.Command(binary, args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()

	return nil
}

func kubectl(args []string) (string, error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return "", err
	}

	// 1. prepare the arguments
	// args := []string{"create", "configmap", name, "--namespace", namespace, fmt.Sprintf("--from-file=%s=%s", name, configFileName)}
	log.Debugf("Exec %s, %v", binary, args)

	env := os.Environ()
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}

	// return syscall.Exec(cmd, args, env)
	// 2. execute the command
	cmd := exec.Command(binary, args...)
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(string(output))
	} else {
		return string(output), nil
	}
}
