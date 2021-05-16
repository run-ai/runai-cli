package util

import (
	"github.com/run-ai/runai-cli/cmd/constants"
	fakeclientset "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned/fake"
	"github.com/run-ai/runai-cli/cmd/util"
	kubeclient "github.com/run-ai/runai-cli/pkg/client"
	prom "github.com/run-ai/runai-cli/pkg/prometheus"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var runaiPodTemplate = v1.PodTemplateSpec{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app": "job-name",
		},
	},
	Spec: v1.PodSpec{
		SchedulerName: constants.SchedulerName,
	},
}

// NewClientForTesting creates a new client for testing purposes
func NewClientForTesting(clientset kubernetes.Interface) *kubeclient.Client {
	client := kubeclient.Client{}
	client.SetClientset(clientset)
	return &client
}

// GetClientWithObject creates a new client with given objects already "created" in its system
func GetClientWithObject(objects []runtime.Object) (kubeclient.Client, *fakeclientset.Clientset) {
	client := fake.NewSimpleClientset(objects...)
	return *NewClientForTesting(client), fakeclientset.NewSimpleClientset()
}

func GetRunaiJob(namespace, jobName, jobUUID string) *batch.Job {
	var labelSelector = make(map[string]string)
	labelSelector["controller-uid"] = jobUUID

	return &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      jobName,
			UID:       types.UID(jobUUID),
			Annotations: map[string]string{
				util.WorkloadCurrentAllocatedGPUsMemory: "10000",
				constants.WorkloadUsedNodes:             "test_node",
				"user":                                  "test_user",
			},
		},
		Spec: batch.JobSpec{
			Template: runaiPodTemplate,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector,
			},
		},
		Status: batch.JobStatus{},
	}
}

func CreatePodOwnedBy(namespace, podName string, labelSelector map[string]string, ownerUUID string, ownerKind string, ownerName string) *v1.Pod {
	controller := true
	if labelSelector == nil {
		labelSelector = make(map[string]string)
	}
	labelSelector["project"] = "test_project"
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels:    labelSelector,
			OwnerReferences: []metav1.OwnerReference{
				{
					UID:        types.UID(ownerUUID),
					Kind:       ownerKind,
					Name:       ownerName,
					Controller: &controller,
				},
			}},
		Spec: v1.PodSpec{
			SchedulerName: constants.SchedulerName,
			Containers: []v1.Container{
				{
					Name:  "container-1",
					Image: "image-1",
				},
			},
		},
	}
}

// PrometheusFakeService returns a fake service for prometheus
func PrometheusFakeService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "",
			Namespace: "monitoring",
			Labels: map[string]string{
				"app": "kube-prometheus-stack-prometheus",
			},
		},
		Spec: v1.ServiceSpec{},
	}
}

// FakePrometheusQueryClient fake prom.QueryClient
type FakePrometheusQueryClient struct {
	metrics prom.MetricResultsByItems
	err     error
}

// GroupMultiQueriesToItems for the fake client will just return the values stored in the fake client
func (fpqc *FakePrometheusQueryClient) GroupMultiQueriesToItems(queryMap map[string]string, labelID string) (prom.MetricResultsByItems, error) {
	return fpqc.metrics, fpqc.err
}

// FakePrometheusClient Creates a fake client to query prometheus
func FakePrometheusClient(metrics prom.MetricResultsByItems, err error) prom.QueryClient {
	return &FakePrometheusQueryClient{metrics: metrics, err: err}
}
