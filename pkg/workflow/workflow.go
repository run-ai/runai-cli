package workflow

import (
	"fmt"
	mpiClient "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned"
	runaiClient "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned"
	"github.com/run-ai/runai-cli/cmd/trainer"
	cmdUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/types"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"os"
	"strconv"

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
*	delete training job with the job name
**/

func getServerConfigMapNameByJob(jobName string, namespaceInfo types.NamespaceInfo, clientset kubernetes.Interface) (string, error) {
	namespace := namespaceInfo.Namespace
	maybeConfigMapNames := []string{jobName, fmt.Sprintf("%s-%s", jobName, "runai"), fmt.Sprintf("%s-%s", jobName, "mpijob")}
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
		return deleteJobResourcesWithoutConfigMap(jobName, namespaceInfo, clientset)
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

	err = clientset.CoreV1().ConfigMaps(namespaceInfo.Namespace).Delete(configMapName, &metav1.DeleteOptions{})
	if err != nil {
		log.Warningf("Delete configmap %s failed due to %v. Please clean it manually", configMapName, err)
		log.Warningf("Please run `kubectl delete -n %s cm %s`", namespace, configMapName)
		return err
	}

	return nil
}

func deleteJobResourcesWithoutConfigMap(jobName string, namespaceInfo types.NamespaceInfo, clientset kubernetes.Interface) error {
	client, err := client.GetClient()
	if err != nil {
		return err
	}

	matchedJobs, err := trainer.GetTrainingJobsByTypeMap(jobName, namespaceInfo.Namespace, client)
	if err != nil {
		return err
	} else if len(matchedJobs) == 0 {
		return cmdUtil.GetJobDoesNotExistsInNamespaceError(jobName, namespaceInfo)
	} else if len(matchedJobs) > 1 {
		return fmt.Errorf("there are multiple jobs named %v", jobName)
	}

	for trainerType, jobToDelete := range matchedJobs {
		jobName := jobToDelete.Name()
		if jobToDelete.Trainer() == trainer.RunaiInteractiveType {
			deleteInteractiveJobResources(jobName, namespaceInfo.Namespace, clientset)
		} else {
			switch trainerType {
			case trainer.MpiTrainerType:
				mpiKubeClient := mpiClient.NewForConfigOrDie(client.GetRestConfig())
				if _, err = mpiKubeClient.KubeflowV1alpha2().MPIJobs(namespaceInfo.Namespace).Get(jobName, metav1.GetOptions{}); err == nil {
					err = mpiKubeClient.KubeflowV1alpha2().MPIJobs(namespaceInfo.Namespace).Delete(jobName, &metav1.DeleteOptions{})
					if err != nil {
						log.Warnf(fmt.Sprintf("Failed to remove mpijob %v, it may be removed manually and not by using Run:AI CLI.", jobName))
					}
				}
			case trainer.DefaultRunaiTrainingType:
				runaiClient := runaiClient.NewForConfigOrDie(client.GetRestConfig())
				if _, err = runaiClient.RunV1().RunaiJobs(namespaceInfo.Namespace).Get(jobName, metav1.GetOptions{}); err == nil {
					err = runaiClient.RunV1().RunaiJobs(namespaceInfo.Namespace).Delete(jobName, metav1.DeleteOptions{})
					if err != nil {
						log.Warnf(fmt.Sprintf("Failed to remove runaijob %v, it may be removed manually and not by using Run:AI CLI.", jobName))
					}
				}
			}
		}
	}

	return nil
}

func deleteInteractiveJobResources(jobName, namespace string, clientset kubernetes.Interface) {
	if _, err := clientset.AppsV1().StatefulSets(namespace).Get(jobName, metav1.GetOptions{}); err == nil {
		err := clientset.AppsV1().StatefulSets(namespace).Delete(jobName, &metav1.DeleteOptions{})
		if err != nil {
			log.Warnf(fmt.Sprintf("Failed to remove statefulSet %v, it may be removed manually and not by using Run:AI CLI.", jobName))
		}
	}
	if _, err := clientset.CoreV1().Services(namespace).Get(jobName, metav1.GetOptions{}); err == nil {
		err = clientset.CoreV1().Services(namespace).Delete(jobName, &metav1.DeleteOptions{})
		if err != nil {
			log.Warnf(fmt.Sprintf("Failed to remove service %v, it may be removed manually and not by using Run:AI CLI.", jobName))
		}
	}
	if _, err := clientset.ExtensionsV1beta1().Ingresses(namespace).Get(jobName, metav1.GetOptions{}); err == nil {
		err = clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(jobName, &metav1.DeleteOptions{})
		if err != nil {
			log.Warnf(fmt.Sprintf("Failed to remove ingress %v, it may be removed manually and not by using Run:AI CLI.", jobName))
		}
	}
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
		return nil, fmt.Errorf("the job %s already exists, please delete it first (use 'runai delete %s')", name, name)
	}

	configMapLabelSelector := getConfigMapLabelSelector(name)
	for i := 0; i < configMapGenerationRetries; i++ {
		existingConfigMaps, err := clientset.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{LabelSelector: configMapLabelSelector})
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
	acceptedConfigMap, err := clientset.CoreV1().ConfigMaps(namespace).Create(&configMap)
	if err != nil {
		return nil, err
	}
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
		fmt.Println("Generate the template on:")
		fmt.Println(jobFiles.template)
		return "", nil
	}
	jobName, err := submitJobInternal(name, namespace, generateSuffix, values, chart, clientset)
	if err != nil {
		return "", err
	}
	return jobName, nil
}
