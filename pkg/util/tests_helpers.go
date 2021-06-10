// +build test

package util

import (
	"fmt"

	"github.com/run-ai/runai-cli/cmd/constants"
	runaijobv1 "github.com/run-ai/runai-cli/cmd/mpi/api/runaijob/v1"
	mpi "github.com/run-ai/runai-cli/cmd/mpi/api/v1alpha2"
	fakeclientset "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned/fake"
	"github.com/run-ai/runai-cli/cmd/util"
	kubeclient "github.com/run-ai/runai-cli/pkg/client"
	prom "github.com/run-ai/runai-cli/pkg/prometheus"
	runaitypes "github.com/run-ai/runai-cli/pkg/types"
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
	client, _ := kubeclient.GetClient()
	client.SetClientset(clientset)
	return client
}

func filterRunAIObjects(objects []runtime.Object, filterIn bool) (filtered []runtime.Object) {
	for _, obj := range objects {
		objGroup := obj.GetObjectKind().GroupVersionKind().Group
		if objGroup == runaijobv1.SchemeGroupVersion.Group || objGroup == mpi.GroupName {
			if filterIn {
				filtered = append(filtered, obj)
			}
		} else if !filterIn {
			filtered = append(filtered, obj)
		}
	}
	return
}

// GetClientWithObject creates a new client with given objects already "created" in its system
func GetClientWithObject(objects []runtime.Object) (kubeclient.Client, *fakeclientset.Clientset) {
	client := fake.NewSimpleClientset(filterRunAIObjects(objects, false)...)
	return *NewClientForTesting(client), fakeclientset.NewSimpleClientset(filterRunAIObjects(objects, true)...)
}

func GetRunaiJob(namespace, jobName, jobUUID string) *runaijobv1.RunaiJob {
	var labelSelector = make(map[string]string)
	labelSelector["controller-uid"] = jobUUID

	return &runaijobv1.RunaiJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       string(runaitypes.ResourceTypeRunaiJob),
			APIVersion: fmt.Sprintf("%s/%s", runaijobv1.GroupName, "v1"),
		},
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
		Spec: runaijobv1.JobSpec{
			Template: runaiPodTemplate,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector,
			},
		},
		Status: runaijobv1.JobStatus{},
	}
}

func GetMPIJob(namespace, jobName, jobUUID string) *mpi.MPIJob {
	var labelSelector = make(map[string]string)
	labelSelector["controller-uid"] = jobUUID

	return &mpi.MPIJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       string(mpi.Kind),
			APIVersion: fmt.Sprintf("%s/%s", mpi.GroupName, mpi.GroupVersion),
		},
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
	}
}

func CreatePodOwnedBy(namespace, podName string, labelSelector map[string]string, ownerUUID string, ownerKind string, ownerName string) *v1.Pod {
	controller := true
	if labelSelector == nil {
		labelSelector = make(map[string]string)
	}
	labelSelector["project"] = "test_project"
	now := metav1.Now()
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
		Status: v1.PodStatus{
			StartTime: &now,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "Pending",
					State: v1.ContainerState{
						Waiting: &v1.ContainerStateWaiting{
							Reason: "Pending",
						},
					},
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
