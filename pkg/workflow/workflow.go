package workflow

import (
	"fmt"
	"os"

	"io/ioutil"

	"github.com/kubeflow/arena/pkg/config"
	"github.com/kubeflow/arena/pkg/util/helm"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

/**
*	delete training job with the job name
**/

func DeleteJob(name, namespace, trainingType string, clientset kubernetes.Interface) error {
	jobName := GetJobName(name, trainingType)

	appInfoFileName, err := kubectl.SaveAppConfigMapToFile(jobName, "app", namespace)
	if err != nil {
		log.Debugf("Failed to SaveAppConfigMapToFile due to %v", err)
		return err
	}

	result, err := kubectl.UninstallAppsWithAppInfoFile(appInfoFileName, namespace)
	log.Debugf("%s", result)
	if err != nil {
		log.Warnf("Failed to remove some of the job's resources, they might have been removed manually and not by using Run:AI CLI.")
	}

	_, err = clientset.CoreV1().ConfigMaps(namespace).Get(jobName, metav1.GetOptions{})

	if err != nil {
		log.Debugf("Skip deletion of ConfigMap %s, because the ConfigMap does not exists.", jobName)
		return nil
	}

	err = kubectl.DeleteAppConfigMap(jobName, namespace)
	if err != nil {
		log.Warningf("Delete configmap %s failed, please clean it manually due to %v.", jobName, err)
		log.Warningf("Please run `kubectl delete -n %s cm %s`", namespace, jobName)
	}

	return nil
}

/**
*	Submit training job
**/

func GetDefaultValuesFile(environmentValues string) (string, error) {
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

func GetJobName(name string, trainingType string) string {
	return fmt.Sprintf("%s-%s", name, trainingType)
}

func SubmitJob(name string, trainingType string, namespace string, values interface{}, environmentValues string, chart string, clientset kubernetes.Interface, dryRun bool) error {
	jobName := GetJobName(name, trainingType)

	if !dryRun {
		found := kubectl.CheckAppConfigMap(fmt.Sprintf("%s-%s", name, trainingType), namespace)
		if found {
			return fmt.Errorf("The job %s already exists, please delete it first. use '%s delete %s'", name, config.CLIName, name)
		}
	}

	// 1. Generate value file
	valueFileName, err := helm.GenerateValueFile(values)
	if err != nil {
		return err
	}

	envValuesFile := ""
	if environmentValues != "" {
		envValuesFile, err = GetDefaultValuesFile(environmentValues)
		if err != nil {
			log.Debugln(err)
			return fmt.Errorf("Error getting default values file of cluster")
		}
	}

	if err != nil {
		log.Debugln(err)
		return fmt.Errorf("Error getting default values file of cluster")
	}

	// 2. Generate Template file
	template, err := helm.GenerateHelmTemplate(name, namespace, valueFileName, envValuesFile, chart)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Println("Generate the template on:")
		fmt.Println(template)
		return nil
	}

	// 3. Generate AppInfo file
	appInfoFileName, err := kubectl.SaveAppInfo(template, namespace)
	if err != nil {
		return err
	}

	// 4. Keep value file in configmap
	chartName := helm.GetChartName(chart)
	chartVersion, err := helm.GetChartVersion(chart)
	if err != nil {
		return err
	}

	err = kubectl.CreateAppConfigmap(jobName,
		namespace,
		valueFileName,
		envValuesFile,
		appInfoFileName,
		chartName,
		chartVersion)
	if err != nil {
		return err
	}
	err = kubectl.LabelAppConfigmap(jobName, namespace, kubectl.JOB_CONFIG_LABEL)
	if err != nil {
		return err
	}

	// 5. Create Application
	_, err = kubectl.UninstallAppsWithAppInfoFile(appInfoFileName, namespace)
	if err != nil {
		log.Debugf("Failed to UninstallAppsWithAppInfoFile due to %v", err)
	}

	result, err := kubectl.InstallApps(template, namespace)
	log.Debugf("%s", result)

	// Clean up because creation of application failed.
	if err != nil {
		log.Warnf("Creation of job failed. Cleaning up...")

		jobName := GetJobName(name, trainingType)
		_, cleanUpErr := kubectl.UninstallAppsWithAppInfoFile(appInfoFileName, namespace)
		if cleanUpErr != nil {
			log.Debugf("Failed to uninstall app with configmap.")
		}
		cleanUpErr = kubectl.DeleteAppConfigMap(jobName, namespace)
		if cleanUpErr != nil {
			log.Debugf("Failed to cleanup configmap %s", jobName)
		}

		return fmt.Errorf("Failed submitting the job:\n %s", err.Error())
	}

	// 6. Clean up the template file
	if log.GetLevel() != log.DebugLevel {
		err = os.Remove(valueFileName)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", valueFileName, err)
		}

		err = os.Remove(template)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", template, err)
		}

		err = os.Remove(appInfoFileName)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", appInfoFileName, err)
		}
	}

	return nil
}
