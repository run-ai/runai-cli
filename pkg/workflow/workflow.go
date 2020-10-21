package workflow

import (
	"fmt"
	cmdUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/types"
	"os"
	"strconv"

	"io/ioutil"

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
	envValuesFile   string
	template        string
	appInfoFileName string
}

/**
*	delete training job with the job name
**/


func getServerConfigMapNameByJob(jobName string, namespaceInfo types.NamespaceInfo, clientset kubernetes.Interface) (string, error) {
	namespace := namespaceInfo.Namespace
	maybeConfigMapNames := []string{jobName, fmt.Sprintf("%s-%s", jobName, "runai"),fmt.Sprintf("%s-%s", jobName, "mpijob")}
	for _, maybeConfigMapName := range maybeConfigMapNames {
		configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(maybeConfigMapName, metav1.GetOptions{})
		if err == nil {
			return configMap.Name, nil
		}
	}
	return "", cmdUtil.GetJobDoesNotExistsInNamespaceError(jobName, namespaceInfo)
}

func DeleteJob(jobName string, namespaceInfo types.NamespaceInfo, clientset kubernetes.Interface) error {
	namespace := namespaceInfo.Namespace
	configMapName, err := getServerConfigMapNameByJob(jobName, namespaceInfo, clientset)
	if err != nil {
		return err
	}

	appInfoFileName, err := kubectl.SaveAppConfigMapToFile(configMapName, "app", namespace)
	if err != nil {
		log.Debugf("Failed to SaveAppConfigMapToFile due to %v", err)
	} else {
		result, err := kubectl.UninstallAppsWithAppInfoFile(appInfoFileName, namespace)
		log.Debugf("%s", result)
		if err != nil {
			log.Warnf("Failed to remove some of the job's resources, they might have been removed manually and not by using Run:AI CLI.")
		}
	}

	err = kubectl.DeleteAppConfigMap(configMapName, namespace)
	if err != nil {
		log.Warningf("Delete configmap %s failed, please clean it manually due to %v.", configMapName, err)
		log.Warningf("Please run `kubectl delete -n %s cm %s`", namespace, configMapName)
		return err
	}

	return nil
}

/**
*	Submit training job
**/

func getDefaultValuesFile(environmentValues string) (string, error) {
	valueFile, err := ioutil.TempFile(os.TempDir(), "values")
	if err != nil {
		return "", err
	}

	_, err = valueFile.WriteString(environmentValues)

	if err != nil {
		return "", err
	}

	log.Debugf("Wrote default cluster values file to path %s", valueFile.Name())

	return valueFile.Name(), nil
}

func generateJobFiles(name string, namespace string, values interface{}, environmentValues string, chart string) (*JobFiles, error) {
	valueFileName, err := helm.GenerateValueFile(values)
	if err != nil {
		return nil, err
	}

	envValuesFile := ""
	if environmentValues != "" {
		envValuesFile, err = getDefaultValuesFile(environmentValues)
		if err != nil {
			log.Debugln(err)
			cleanupSingleFile(valueFileName)
			return nil, fmt.Errorf("Error getting default values file of cluster")
		}
	}

	// 2. Generate Template file
	template, err := helm.GenerateHelmTemplate(name, namespace, valueFileName, envValuesFile, chart)
	if err != nil {
		cleanupSingleFile(environmentValues)
		cleanupSingleFile(valueFileName)
		return nil, err
	}

	// 3. Generate AppInfo file
	appInfoFileName, err := kubectl.SaveAppInfo(template, namespace)
	if err != nil {
		cleanupSingleFile(template)
		cleanupSingleFile(environmentValues)
		cleanupSingleFile(valueFileName)
		return nil, err
	}


	jobFiles := &JobFiles{
		valueFileName:   valueFileName,
		envValuesFile:   envValuesFile,
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

func getConfigMapName(name string, index int) string {
	if index == 0 {
		return name
	}
	return fmt.Sprintf("%s-%d", name, index)
}

func submitConfigMap(name, namespace string, generateName bool, clientset kubernetes.Interface) (*corev1.ConfigMap, error) {
	maybeConfigMapName := getConfigMapName(name, 0)

	configMap, err := createEmptyConfigMap(name, name, namespace, 0, clientset)
	if err == nil {
		return configMap, nil
	}

	if !generateName {
		return nil, fmt.Errorf("the job %s already exists, either delete it first (use 'runai delete <job-name>' ) or submit the job again using the flag --generate-name", name)
	}

	configMapLabelSelector := getConfigMapLabelSelector(maybeConfigMapName)
	for i := 0; i < configMapGenerationRetries; i ++ {
		existingConfigMaps, err := clientset.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{LabelSelector: configMapLabelSelector})
		if err != nil {
			return nil, err
		}
		configMapIndex := getSmallestUnoccupiedIndex(existingConfigMaps.Items)
		maybeConfigMapName = getConfigMapName(name, configMapIndex)

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
	acceptedConfigMap, err := clientset.CoreV1().ConfigMaps(namespace).Create(&configMap)
	if err != nil {
		return nil, err
	}
	return acceptedConfigMap, nil
}

func populateConfigMap(configMap *corev1.ConfigMap, chartName, chartVersion, envValuesFile, valuesFileName, appInfoFileName, namespace string, clientset kubernetes.Interface) error {
	data := make(map[string]string)
	data[chartName] = chartVersion
	if envValuesFile != "" {
		envFileContent, err := ioutil.ReadFile(envValuesFile)
		if err != nil {
			return err
		}
		data["env-values"] = string(envFileContent)
	}
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
	_, err = clientset.CoreV1().ConfigMaps(namespace).Update(configMap)
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
	cleanupSingleFile(files.envValuesFile)
}

func submitJobInternal(name, namespace string, generateName bool, values interface{}, environmentValues string, chart string, clientset kubernetes.Interface) (string, error) {
	configMap, err := submitConfigMap(name, namespace, generateName, clientset)
	if err != nil {
		return "", err
	}
	jobName := configMap.Name
	jobFiles, err := generateJobFiles(jobName, namespace, values, environmentValues, chart)
	if err != nil {
		return jobName, err
	}
	defer cleanupJobFiles(jobFiles)

	chartName := helm.GetChartName(chart)
	chartVersion, err := helm.GetChartVersion(chart)
	if err != nil {
		return jobName, err
	}

	err = populateConfigMap(configMap, chartName, chartVersion, jobFiles.envValuesFile, jobFiles.valueFileName, jobFiles.appInfoFileName, namespace, clientset)
	if err != nil {
		return jobName, err
	}

	_, err = kubectl.InstallApps(jobFiles.template, namespace)
	if err != nil {
		return jobName, err
	}
	return jobName, nil
}

func SubmitJob(name, namespace string, generateName bool, values interface{}, environmentValues string, chart string, clientset kubernetes.Interface, dryRun bool) (string, error) {
	if dryRun {
		jobFiles, err := generateJobFiles(name, namespace, values, environmentValues, chart)
		if err != nil {
			return "", err
		}
		fmt.Println("Generate the template on:")
		fmt.Println(jobFiles.template)
		return "", nil
	}
	jobName, err := submitJobInternal(name, namespace, generateName, values, environmentValues, chart, clientset)
	if err != nil {
		return "", err
	}
	return jobName, nil
}