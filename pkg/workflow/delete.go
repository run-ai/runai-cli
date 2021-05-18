package workflow

//    *NOTE*   The functionality of this code has been replaced by calling to researcher-service API.
//    *NOTE*   It is currently left for compatibility, but expected to be removed

import (
	"fmt"
	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	mpiClient "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned"
	runaiClient "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned"

	"github.com/run-ai/runai-cli/cmd/trainer"
	cmdUtil "github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/util/kubectl"
)

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

	err = clientset.CoreV1().ConfigMaps(namespaceInfo.Namespace).Delete(configMapName, &metav1.DeleteOptions{})
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
		err = clientset.AppsV1().Deployments(namespaceInfo.Namespace).Delete(jobName, &metav1.DeleteOptions{})
	case string(types.ResourceTypeJob):
		err = clientset.BatchV1().Jobs(namespaceInfo.Namespace).Delete(jobName, &metav1.DeleteOptions{})
	case string(types.ResourceTypeStatefulSet):
		err = clientset.AppsV1().StatefulSets(namespaceInfo.Namespace).Delete(jobName, &metav1.DeleteOptions{})
	case string(types.ResourceTypeRunaiJob):
		runaiClient := runaiClient.NewForConfigOrDie(client.GetRestConfig())
		err = runaiClient.RunV1().RunaiJobs(namespaceInfo.Namespace).Delete(jobName, metav1.DeleteOptions{})
	case string(types.ResourceTypePod):
		err = clientset.CoreV1().Pods(namespaceInfo.Namespace).Delete(jobName, &metav1.DeleteOptions{})
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
	err := clientset.CoreV1().Services(namespace).Delete(jobName, &metav1.DeleteOptions{})
	if err != nil {
		log.Debugf("Failed to remove service %v, it may be removed manually and not by using Run:AI CLI.", jobName)
	}
	err = clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(jobName, &metav1.DeleteOptions{})
	if err != nil {
		log.Debugf("Failed to remove ingress %v, it may be removed manually and not by using Run:AI CLI.", jobName)
	}
	err = clientset.CoreV1().ConfigMaps(namespace).Delete(jobName, &metav1.DeleteOptions{})
	if err != nil {
		log.Debugf("Failed to remove configmap %s failed due to %v. Please clean it manually\n", jobName, err)
	}
}
