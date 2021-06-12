package workflow

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/run-ai/runai-cli/pkg/util/helm"
	"github.com/run-ai/runai-cli/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	BaseNameLabelSelectorName  = "base-name"
	baseIndexLabelSelectorName = "base-name-index"
	configMapGenerationRetries = 5
)

type JobFiles struct {
	valueFileName   string
	template        string
	appInfoFileName string
}

/**
*	Submit training job
**/

func generateJobFiles(name string, namespace string, values interface{}, chart string) (*JobFiles, error) {
	valueFileName, err := helm.GenerateValueFile(values)
	if err != nil {
		return nil, err
	}

	// 2. Generate Template file
	template, err := helm.GenerateHelmTemplate(name, namespace, valueFileName, chart)
	if err != nil {
		cleanupSingleFile(valueFileName)
		return nil, err
	}

	// 3. Generate AppInfo file
	appInfoFileName, err := kubectl.SaveAppInfo(template, namespace)
	if err != nil {
		cleanupSingleFile(template)
		cleanupSingleFile(valueFileName)
		return nil, err
	}

	jobFiles := &JobFiles{
		valueFileName:   valueFileName,
		template:        template,
		appInfoFileName: appInfoFileName,
	}

	return jobFiles, nil

}

func getConfigMapLabelSelector(configMapName string) string {
	return fmt.Sprintf("%s=%s", BaseNameLabelSelectorName, configMapName)
}

func getSmallestUnoccupiedIndex(configMaps []corev1.ConfigMap) int {
	occupationMap := make(map[string]bool)
	for _, configMap := range configMaps {
		occupationMap[configMap.Labels[baseIndexLabelSelectorName]] = true
	}

	for i := 1; i < len(configMaps); i++ {
		if !occupationMap[strconv.Itoa(i)] {
			return i
		}
	}

	return len(configMaps)
}

func getConfigMapName(name string, index int, generateSuffix bool) string {
	if !generateSuffix {
		return name
	}
	return fmt.Sprintf("%s-%d", name, index)
}

func submitConfigMap(name, namespace string, generateSuffix bool, clientset kubernetes.Interface) (*corev1.ConfigMap, error) {
	maybeConfigMapName := getConfigMapName(name, 0, generateSuffix)

	configMap, err := createEmptyConfigMap(maybeConfigMapName, name, namespace, 0, clientset)
	if err == nil {
		return configMap, nil
	}

	if apiErr, ok := err.(*errors.StatusError); ok && (apiErr.ErrStatus.Code == 403 || apiErr.ErrStatus.Code == 401) {
		return nil, err
	}

	if !generateSuffix {
		log.Debugf("Failed to create job name: <%v>, error: <%v>", name, err)
		return nil, fmt.Errorf("the job %s already exists, please delete it first (use 'runai delete %s')", name, name)
	}

	configMapLabelSelector := getConfigMapLabelSelector(name)
	for i := 0; i < configMapGenerationRetries; i++ {
		existingConfigMaps, err := clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: configMapLabelSelector})
		if err != nil {
			return nil, err
		}
		configMapIndex := getSmallestUnoccupiedIndex(existingConfigMaps.Items)
		maybeConfigMapName = getConfigMapName(name, configMapIndex, generateSuffix)
		configMap, err = createEmptyConfigMap(maybeConfigMapName, name, namespace, configMapIndex, clientset)
		if err == nil {
			return configMap, nil
		}
	}

	return nil, fmt.Errorf("job creation has failed. Please try again")
}

func createEmptyConfigMap(name, baseName, namespace string, index int, clientset kubernetes.Interface) (*corev1.ConfigMap, error) {
	labels := make(map[string]string)
	labels[kubectl.JOB_CONFIG_LABEL_KEY] = kubectl.JOB_CONFIG_LABEL_VALUES
	labels[baseIndexLabelSelectorName] = strconv.Itoa(index)
	labels[BaseNameLabelSelectorName] = baseName

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
	acceptedConfigMap, err := clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &configMap, metav1.CreateOptions{})
	if err != nil {
		log.Debugf("Failed to create configmap name: <%v>, error: <%v>", name, err)
		return nil, err
	}
	log.Debugf("Create configmap name: <%v>", name)
	return acceptedConfigMap, nil
}

func populateConfigMap(configMap *corev1.ConfigMap, chartName, chartVersion, valuesFileName, appInfoFileName, namespace string, clientset kubernetes.Interface) error {
	data := make(map[string]string)
	data[chartName] = chartVersion
	valuesFileContent, err := ioutil.ReadFile(valuesFileName)
	if err != nil {
		return err
	}
	data["values"] = string(valuesFileContent)
	appFileContent, err := ioutil.ReadFile(appInfoFileName)
	if err != nil {
		return err
	}

	data["app"] = string(appFileContent)

	configMap.Data = data
	_, err = clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions {})
	return err
}

func cleanupSingleFile(file string) {
	if _, err := os.Stat(file); err == nil {
		err = os.Remove(file)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", file, err)
		}
	}
}

func cleanupJobFiles(files *JobFiles) {
	cleanupSingleFile(files.valueFileName)
	cleanupSingleFile(files.template)
	cleanupSingleFile(files.appInfoFileName)
}

func submitJobInternal(name, namespace string, generateSuffix bool, values interface{}, chart string, clientset kubernetes.Interface) (string, error) {
	configMap, err := submitConfigMap(name, namespace, generateSuffix, clientset)
	if err != nil {
		return "", err
	}
	jobName := configMap.Name
	jobFiles, err := generateJobFiles(jobName, namespace, values, chart)
	if err != nil {
		return jobName, err
	}
	defer cleanupJobFiles(jobFiles)

	chartName := helm.GetChartName(chart)
	chartVersion, err := helm.GetChartVersion(chart)
	if err != nil {
		return jobName, err
	}

	err = populateConfigMap(configMap, chartName, chartVersion, jobFiles.valueFileName, jobFiles.appInfoFileName, namespace, clientset)
	if err != nil {
		return jobName, err
	}

	_, err = kubectl.UninstallAppsWithAppInfoFile(jobFiles.appInfoFileName, namespace)
	if err != nil {
		log.Debugf("Failed to UninstallAppsWithAppInfoFile due to %v", err)
	}
	_, err = kubectl.InstallApps(jobFiles.template, namespace)
	if err != nil {
		return jobName, err
	}
	return jobName, nil
}

func SubmitJob(name, namespace string, generateSuffix bool, values interface{}, chart string, clientset kubernetes.Interface, dryRun bool) (string, error) {
	if dryRun {
		jobFiles, err := generateJobFiles(name, namespace, values, chart)
		if err != nil {
			return "", err
		}
		fmt.Println("Template YAML file can be found at:")
		fmt.Println(jobFiles.template)
		return "", nil
	}
	jobName, err := submitJobInternal(name, namespace, generateSuffix, values, chart, clientset)
	if err != nil {
		return "", err
	}
	return jobName, nil
}
