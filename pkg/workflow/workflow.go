package workflow

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	mpiClient "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned"
	runaiClient "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned"
	"github.com/run-ai/runai-cli/cmd/trainer"
	cmdUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/types"
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
*	delete training job with the job name
**/

func getServerConfigMapNameByJob(jobName string, namespaceInfo types.NamespaceInfo, clientset kubernetes.Interface) (string, error) {
	namespace := namespaceInfo.Namespace
	maybeConfigMapNames := []string{jobName, fmt.Sprintf("%s-%s", jobName, "runai"), fmt.Sprintf("%s-%s", jobName, "mpijob")}
	for _, maybeConfigMapName := range maybeConfigMapNames {
		configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), maybeConfigMapName, metav1.GetOptions{})
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
		log.Debugf("Failed to find configmap for, error: %v\n", err)
		return deleteJobResourcesWithoutConfigMap(jobName, namespaceInfo, clientset)
	}

	appInfoFileName, err := kubectl.SaveAppConfigMapToFile(configMapName, "app", namespace)
	if err != nil {
		log.Debugf("Failed to SaveAppConfigMapToFile due to %v\n", err)
		return deleteJobResourcesWithoutConfigMap(jobName, namespaceInfo, clientset)
	}

	result, err := kubectl.UninstallAppsWithAppInfoFile(appInfoFileName, namespace)
	log.Debugf("%s", result)
	if err != nil {
		log.Debugf("Failed to remove some of the job's resources, they might have been removed manually and not by using Run:AI CLI.\n")
		return deleteJobResourcesWithoutConfigMap(jobName, namespaceInfo, clientset)
	}

	err = clientset.CoreV1().ConfigMaps(namespaceInfo.Namespace).Delete(context.TODO(), configMapName, metav1.DeleteOptions{})
	if err != nil {
		log.Debugf("Delete configmap %s failed due to %v. Please clean it manually\n", configMapName, err)
		log.Debugf("Please run `kubectl delete -n %s cm %s\n`", namespace, configMapName)
		return err
	}
	deletedJobMessage(jobName)

	return nil
}

func deletedJobMessage(jobName string) {
	fmt.Printf("Successfully deleted job: %s\n", jobName)
}

func deleteJobResourcesWithoutConfigMap(jobName string, namespaceInfo types.NamespaceInfo, clientset kubernetes.Interface) error {
	client, err := client.GetClient()
	if err != nil {
		return err
	}
	jobToDelete, err := trainer.SearchTrainingJob(client, jobName, "", namespaceInfo)
	if err != nil {
		deleteAdditionalJobResources(jobName, namespaceInfo.Namespace, clientset)
		return cmdUtil.GetJobDoesNotExistsInNamespaceError(jobName, namespaceInfo)
	}

	switch jobToDelete.WorkloadType() {
	case string(types.MpiWorkloadType):
		mpiKubeClient := mpiClient.NewForConfigOrDie(client.GetRestConfig())
		err = mpiKubeClient.KubeflowV1alpha2().MPIJobs(namespaceInfo.Namespace).Delete(jobName, &metav1.DeleteOptions{})
	case string(types.ResourceTypeDeployment):
		err = clientset.AppsV1().Deployments(namespaceInfo.Namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{})
	case string(types.ResourceTypeJob):
		err = clientset.BatchV1().Jobs(namespaceInfo.Namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{})
	case string(types.ResourceTypeStatefulSet):
		err = clientset.AppsV1().StatefulSets(namespaceInfo.Namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{})
	case string(types.ResourceTypeRunaiJob):
		runaiClient := runaiClient.NewForConfigOrDie(client.GetRestConfig())
		err = runaiClient.RunV1().RunaiJobs(namespaceInfo.Namespace).Delete(jobName, metav1.DeleteOptions{})
	case string(types.ResourceTypePod):
		err = clientset.CoreV1().Pods(namespaceInfo.Namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{})
	default:
		log.Warningf("Unexpected type for job, type: %v\n", jobToDelete.WorkloadType())
	}
	if err != nil {
		log.Debugf("Failed to remove job %v, it may be removed manually and not by using Run:AI CLI.\n", jobName)
	}

	deleteAdditionalJobResources(jobName, namespaceInfo.Namespace, clientset)
	if err != nil {
		deletedJobMessage(jobName)
	}

	return nil
}

func deleteAdditionalJobResources(jobName, namespace string, clientset kubernetes.Interface) {
	err := clientset.CoreV1().Services(namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{})
	if err != nil {
		log.Debugf("Failed to remove service %v, it may be removed manually and not by using Run:AI CLI.", jobName)
	}
	err = clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{})
	if err != nil {
		log.Debugf("Failed to remove ingress %v, it may be removed manually and not by using Run:AI CLI.", jobName)
	}
	err = clientset.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{})
	if err != nil {
		log.Debugf("Failed to remove configmap %s failed due to %v. Please clean it manually\n", jobName, err)
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
