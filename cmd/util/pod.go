package util

import (
	"fmt"
	"github.com/run-ai/runai-cli/cmd/trainer"

	// "os"
	"time"

	"github.com/run-ai/runai-cli/pkg/client"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NotReadyPodTimeoutMsg = "Timeout waiting for job to start running"
	NotReadyPodWaitingMsg = "Waiting for pod to start running..."
)

// GetPod by its name and namespace
func GetPod(name, namespace string) (*v1.Pod, error) {
	client, err := client.GetClient()
	if err != nil {
		return nil, err
	}
	return client.GetClientset().CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
}

// WaitForPod waiting to the pod
func WaitForPod(podName, podNamespace, waitingMsg string, timeout time.Duration, timeoutMsg string, exitCondition func(*v1.Pod) (bool, error)) (pod *v1.Pod, err error) {
	shouldStopAt := time.Now().Add(timeout)

	for i, exit := 0, false; ; i++ {
		pod, err = GetPod(podName, podNamespace)
		if err != nil {
			return
		}

		exit, err = exitCondition(pod)
		if err != nil || exit {
			return
		}

		if shouldStopAt.Before(time.Now()) {
			return nil, fmt.Errorf(timeoutMsg)
		}

		if i == 0 && len(waitingMsg) != 0 {
			fmt.Println(waitingMsg)
		}

		time.Sleep(time.Second)
	}
}

func WaitForPodCreation(podName string, job trainer.TrainingJob, timeout time.Duration) (pod *v1.Pod, err error) {
	shouldStopAt := time.Now().Add(timeout)

	for true {
		if len(podName) == 0 {
			pod = job.ChiefPod()
		} else {
			pods := job.AllPods()
			for _, p := range pods {
				if podName == p.Name {
					pod = &p
					break
				}
			}
		}

		if pod != nil {
			return pod, nil
		}

		if shouldStopAt.Before(time.Now()) {
			return nil, fmt.Errorf("Failed to find pod: '%s' of job: '%s'", podName, job.Name())
		}

		time.Sleep(time.Second)
	}
	return
}

// PodRunning check if the pod is running and ready
func PodRunning(pod *v1.Pod) (bool, error) {
	phase := pod.Status.Phase

	switch phase {
	case v1.PodPending:
		break
	case v1.PodRunning:
		conditions := pod.Status.Conditions
		if conditions == nil {
			return false, nil
		}
		for i := range conditions {
			if conditions[i].Type == corev1.PodReady &&
				conditions[i].Status == corev1.ConditionTrue {
				return true, nil
			}
		}

	default:
		return false, fmt.Errorf("Can't connect to the pod: %s in phase: %s", pod.Name, phase)
	}

	return false, nil
}
