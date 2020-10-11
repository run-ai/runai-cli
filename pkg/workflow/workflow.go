package workflow

import (
	"fmt"
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

/**
*	delete training job with the job name
**/

const (
	submitNewNameRetries = 3
		JobFamilyName = "job-family-name"
		JobFamilyIndex = "job-family-index"
		JobFamilyRoot = "job-family-root")

type getJobSuffixFunc func(name, namespace string) (int, error)

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
		log.Debugf("Skip deletion of ConfigMap %s, because the ConfigMap does not exist.", jobName)
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

type JobFiles struct {
	valueFileName   string
	envValuesFile   string
	template        string
	appInfoFileName string
}

func generateJobFiles(name string, namespace string, values interface{}, environmentValues string, chart string) (*JobFiles, error) {
	valueFileName, err := helm.GenerateValueFile(values)
	if err != nil {
		return nil, err
	}

	envValuesFile := ""
	if environmentValues != "" {
		envValuesFile, err = GetDefaultValuesFile(environmentValues)
		if err != nil {
			log.Debugln(err)
			return nil, fmt.Errorf("Error getting default values file of cluster")
		}
	}

	if err != nil {
		log.Debugln(err)
		return nil, fmt.Errorf("Error getting default values file of cluster")
	}

	// 2. Generate Template file
	template, err := helm.GenerateHelmTemplate(name, namespace, valueFileName, envValuesFile, chart)
	if err != nil {
		return nil, err
	}

	// 3. Generate AppInfo file
	appInfoFileName, err := kubectl.SaveAppInfo(template, namespace)
	if err != nil {
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

func generateJobFilesWithNewName(name *string , namespace, environmentValues, chart string, values interface{}, labels *map[string]string, getJobSuffixFunc getJobSuffixFunc) (*JobFiles, error) {
	jobSuffix, err := getJobSuffixFunc(*name, namespace)
	if err != nil {
		return nil, err
	}
	initialName := *name
	(*labels)[JobFamilyRoot] = strconv.FormatBool(false)

	for i := 0; i < submitNewNameRetries; i++ {
		jobSuffixStr := strconv.Itoa(jobSuffix)
		fmt.Println(jobSuffixStr)
		currentName := initialName + "-" + jobSuffixStr

		(*labels)[JobFamilyIndex] = jobSuffixStr
		generatedJobFiles, err := generateJobFiles(currentName, namespace, values, environmentValues, chart)
		if err != nil {
			return nil, err
		}

		jobExists, err := kubectl.CheckIfAppInfofileContentsExists(generatedJobFiles.appInfoFileName, namespace)

		if !jobExists {
			*name = currentName
			return generatedJobFiles, nil
		}
	}

	return nil, fmt.Errorf("could not submit %s. Please try again", initialName)
}

func SubmitJob(baseName *string, trainingType string, namespace string, values interface{}, labels *map[string]string, environmentValues string, chart string, clientset kubernetes.Interface, getJobSuffixFunc getJobSuffixFunc, dryRun bool) error {
	name := *baseName
	jobName := GetJobName(name, trainingType)
	(*labels)[JobFamilyName] = name
	(*labels)[JobFamilyRoot] = strconv.FormatBool(true)

	var jobFiles *JobFiles

	if !dryRun {
		found, _ := clientset.CoreV1().ConfigMaps(namespace).Get(jobName, metav1.GetOptions{})

		if found != nil && found.Name != "" {
			generatedJobFiles, err := generateJobFiles(name, namespace, values, environmentValues, chart)
			if err != nil {
				return err
			}

			jobFiles = generatedJobFiles

			jobExists, err := kubectl.CheckIfAppInfofileContentsExists(jobFiles.appInfoFileName, namespace)

			if err != nil {
				return err
			}

			if jobExists {
				jobFiles, err = generateJobFilesWithNewName(baseName, namespace, environmentValues, chart, values, labels, getJobSuffixFunc)
				if err != nil {
					return err
				}
				name = *baseName
				jobName = GetJobName(name, trainingType)
			} else {
				// Delete the configmap of the job and continue for the creation of the new one.

				log.Debugf("Configmap for job exists but job itself does not for job %s on namespace %s. Deleting the configmap", name, namespace)
				err := clientset.CoreV1().ConfigMaps(namespace).Delete(jobName, &metav1.DeleteOptions{})

				if err != nil {
					log.Debugf("Could not delete configmap for job %s on namespace %s", name, namespace)
					return fmt.Errorf("Error submitting the job.")
				}
			}
		}
	}

	// Create job files only if did not create them yet
	if jobFiles == nil {
		generatedJobFiles, err := generateJobFiles(name, namespace, values, environmentValues, chart)
		if err != nil {
			return err
		}

		jobFiles = generatedJobFiles
	}

	if dryRun {
		fmt.Println("Generate the template on:")
		fmt.Println(jobFiles.template)
		return nil
	}

	// 4. Keep value file in configmap
	chartName := helm.GetChartName(chart)
	chartVersion, err := helm.GetChartVersion(chart)
	if err != nil {
		return err
	}

	err = createConfigMap(
		jobName,
		namespace,
		jobFiles.valueFileName,
		jobFiles.envValuesFile,
		jobFiles.appInfoFileName,
		chartName,
		chartVersion,
		clientset,
	)
	if err != nil {
		return err
	}

	// 5. Create Application
	_, err = kubectl.UninstallAppsWithAppInfoFile(jobFiles.appInfoFileName, namespace)
	if err != nil {
		log.Debugf("Failed to UninstallAppsWithAppInfoFile due to %v", err)
	}

	result, err := kubectl.InstallApps(jobFiles.template, namespace)
	log.Debugf("%s", result)

	// Clean up because creation of application failed.
	if err != nil {
		log.Warnf("Creation of job failed. Cleaning up...")

		jobName := GetJobName(name, trainingType)
		_, cleanUpErr := kubectl.UninstallAppsWithAppInfoFile(jobFiles.appInfoFileName, namespace)
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
		err = os.Remove(jobFiles.valueFileName)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", jobFiles.valueFileName, err)
		}

		err = os.Remove(jobFiles.template)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", jobFiles.template, err)
		}

		err = os.Remove(jobFiles.appInfoFileName)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", jobFiles.appInfoFileName, err)
		}
	}

	return nil
}

func createConfigMap(jobName string,
	namespace string,
	valueFileName string,
	envValuesFile string,
	appInfoFileName string,
	chartName string,
	chartVersion string, clientset kubernetes.Interface) error {
	lables := make(map[string]string)
	data := make(map[string]string)
	data["app"] = appInfoFileName
	data[chartName] = chartVersion
	if envValuesFile != "" {
		envFileContent, err := ioutil.ReadFile(envValuesFile)
		if err != nil {
			return err
		}

		data["env-values"] = string(envFileContent)
	}

	valuesFileContent, err := ioutil.ReadFile(valueFileName)
	if err != nil {
		return err
	}

	data["values"] = string(valuesFileContent)

	appFileContent, err := ioutil.ReadFile(appInfoFileName)
	if err != nil {
		return err
	}

	data["app"] = string(appFileContent)

	lables[kubectl.JOB_CONFIG_LABEL_KEY] = kubectl.JOB_CONFIG_LABEL_VALUES
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   jobName,
			Labels: lables,
		},
		Data: data,
	}

	_, err = clientset.CoreV1().ConfigMaps(namespace).Create(&configMap)
	return err
}
