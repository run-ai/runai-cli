package trainer

import (
	"testing"

	"github.com/run-ai/runai-cli/cmd/constants"
	fakeclientset "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned/fake"
	kubeclient "github.com/run-ai/runai-cli/pkg/client"
	cmdTypes "github.com/run-ai/runai-cli/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	NAMESPACE string = "default"
)

var runaiPodTemplate = v1.PodTemplateSpec{
	Spec: v1.PodSpec{
		SchedulerName: constants.SchedulerName,
	},
}

func getClientWithObject(objects []runtime.Object) (kubeclient.Client, *fakeclientset.Clientset) {
	client := fake.NewSimpleClientset(objects...)
	return *kubeclient.NewClientForTesting(client), fakeclientset.NewSimpleClientset()
}

func getRunaiReplicaSet() *appsv1.ReplicaSet {
	jobName := "job-name"
	jobUUID := "id1"

	labelSelector := make(map[string]string)

	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NAMESPACE,
			Name:      jobName,
			UID:       types.UID(jobUUID),
		},
		Spec: appsv1.ReplicaSetSpec{
			Template: runaiPodTemplate,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector,
			},
		},
	}
}

func getRunaiStatefulSet() *appsv1.StatefulSet {
	jobName := "job-name"
	jobUUID := "id1"

	labelSelector := make(map[string]string)

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NAMESPACE,
			Name:      jobName,
			UID:       types.UID(jobUUID),
		},
		Spec: appsv1.StatefulSetSpec{
			Template: runaiPodTemplate,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector,
			},
		},
		Status: appsv1.StatefulSetStatus{},
	}
}

func getRunaiJob() *batch.Job {
	jobName := "job-name"
	jobUUID := "id1"

	var labelSelector = make(map[string]string)
	labelSelector["controller-uid"] = jobUUID

	return &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NAMESPACE,
			Name:      jobName,
			UID:       types.UID(jobUUID),
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

func TestJobInclusionInResourcesListCommand(t *testing.T) {
	job := getRunaiJob()

	pod := createPodOwnedBy("pod", nil, string(job.UID), string(cmdTypes.ResourceTypeJob), job.Name)

	objects := []runtime.Object{pod, job}
	kubeClient, runaiclient := getClientWithObject(objects)
	trainer := RunaiTrainer{runaijobClient: runaiclient, client: kubeClient.GetClientset()}
	jobs, _ := trainer.ListTrainingJobs(NAMESPACE)

	trainJob := jobs[0]
	resources := trainJob.Resources()

	if !testResourceIncluded(resources, job.Name, cmdTypes.ResourceTypeJob) {
		t.Errorf("Could not find related job in training job resources")
	}
}

func TestDontListNonRunaiJobs(t *testing.T) {
	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NAMESPACE,
			Name:      "name",
			UID:       types.UID("jobUUID"),
		},
		Spec: batch.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					SchedulerName: "some-scheduler",
				},
			},
		},
	}

	objects := []runtime.Object{job}
	kubeClient, runaiclient := getClientWithObject(objects)
	trainer := RunaiTrainer{runaijobClient: runaiclient, client: kubeClient.GetClientset()}

	jobs, _ := trainer.ListTrainingJobs(NAMESPACE)

	if len(jobs) != 0 {
		t.Errorf("Got too many resources from list command")
	}
}

func TestJobInclusionInResourcesGetCommand(t *testing.T) {
	job := getRunaiJob()

	objects := []runtime.Object{job}
	kubeClient, runaiclient := getClientWithObject(objects)
	trainer := RunaiTrainer{runaijobClient: runaiclient, client: kubeClient.GetClientset()}

	trainJob, _ := trainer.GetTrainingJob(job.Name, NAMESPACE)

	resources := trainJob.Resources()

	if !testResourceIncluded(resources, job.Name, cmdTypes.ResourceTypeJob) {
		t.Errorf("Could not find related job in training job resources")
	}
}

func TestStatefulSetInclusionInResourcesGetCommand(t *testing.T) {
	job := getRunaiStatefulSet()

	objects := []runtime.Object{job}
	kubeClient, runaiclient := getClientWithObject(objects)
	trainer := RunaiTrainer{runaijobClient: runaiclient, client: kubeClient.GetClientset()}

	trainJob, _ := trainer.GetTrainingJob(job.Name, NAMESPACE)

	resources := trainJob.Resources()

	if !testResourceIncluded(resources, job.Name, cmdTypes.ResourceTypeStatefulSet) {
		t.Errorf("Could not find related job in training job resources")
	}
}

func TestReplicaSetInclusionInResourcesGetCommand(t *testing.T) {
	job := getRunaiReplicaSet()

	objects := []runtime.Object{job}
	kubeClient, runaiclient := getClientWithObject(objects)
	trainer := RunaiTrainer{runaijobClient: runaiclient, client: kubeClient.GetClientset()}

	trainJob, _ := trainer.GetTrainingJob(job.Name, NAMESPACE)

	resources := trainJob.Resources()

	if !testResourceIncluded(resources, job.Name, cmdTypes.ResourceTypeReplicaset) {
		t.Errorf("Could not find related job in training job resources")
	}
}

func TestIncludeMultiplePodsInReplicaset(t *testing.T) {
	job := getRunaiReplicaSet()

	pod1 := createPodOwnedBy("pod1", job.Spec.Selector.MatchLabels, string(job.UID), string(cmdTypes.ResourceTypeJob), job.Name)
	pod2 := createPodOwnedBy("pod2", job.Spec.Selector.MatchLabels, string(job.UID), string(cmdTypes.ResourceTypeJob), job.Name)

	objects := []runtime.Object{job, pod1, pod2}
	kubeClient, runaiclient := getClientWithObject(objects)
	trainer := RunaiTrainer{runaijobClient: runaiclient, client: kubeClient.GetClientset()}

	trainJob, _ := trainer.GetTrainingJob(job.Name, NAMESPACE)

	if len(trainJob.AllPods()) != 2 {
		t.Errorf("Did not get all pod owned by job")
	}
}

func TestIncludeMultiplePodsInStatefulset(t *testing.T) {
	job := getRunaiStatefulSet()

	pod1 := createPodOwnedBy("pod1", job.Spec.Selector.MatchLabels, string(job.UID), string(cmdTypes.ResourceTypeJob), job.Name)
	pod2 := createPodOwnedBy("pod2", job.Spec.Selector.MatchLabels, string(job.UID), string(cmdTypes.ResourceTypeJob), job.Name)

	objects := []runtime.Object{job, pod1, pod2}
	kubeClient, runaiclient := getClientWithObject(objects)
	trainer := RunaiTrainer{runaijobClient: runaiclient, client: kubeClient.GetClientset()}

	trainJob, _ := trainer.GetTrainingJob(job.Name, NAMESPACE)

	if len(trainJob.AllPods()) != 2 {
		t.Errorf("Did not get all pod owned by job")
	}
}

func TestIncludeMultiplePodsInJob(t *testing.T) {
	job := getRunaiJob()

	pod1 := createPodOwnedBy("pod1", job.Spec.Selector.MatchLabels, string(job.UID), string(cmdTypes.ResourceTypeJob), job.Name)
	pod2 := createPodOwnedBy("pod2", job.Spec.Selector.MatchLabels, string(job.UID), string(cmdTypes.ResourceTypeJob), job.Name)

	objects := []runtime.Object{job, pod1, pod2}
	kubeClient, runaiclient := getClientWithObject(objects)
	trainer := RunaiTrainer{runaijobClient: runaiclient, client: kubeClient.GetClientset()}

	trainJob, _ := trainer.GetTrainingJob(job.Name, NAMESPACE)

	if len(trainJob.AllPods()) != 2 {
		t.Errorf("Did not get all pod owned by job")
	}
}

func TestDontGetNotRunaiJob(t *testing.T) {
	job := getRunaiJob()
	job.Spec.Template.Spec.SchedulerName = "some-scheduler"

	objects := []runtime.Object{job}
	kubeClient, runaiclient := getClientWithObject(objects)
	trainer := RunaiTrainer{runaijobClient: runaiclient, client: kubeClient.GetClientset()}

	trainJob, _ := trainer.GetTrainingJob(job.Name, NAMESPACE)

	if trainJob != nil {
		t.Errorf("Expected nil as return, but got a job")
	}
}

func TestStatefulsetJobIsInteractive(t *testing.T) {
	job := getRunaiStatefulSet()

	objects := []runtime.Object{job}
	kubeClient, runaiclient := getClientWithObject(objects)
	trainer := RunaiTrainer{runaijobClient: runaiclient, client: kubeClient.GetClientset()}

	jobs, _ := trainer.ListTrainingJobs(NAMESPACE)

	jobType := jobs[0].Trainer()
	if jobType != "Interactive" {
		t.Errorf("Expected job to be interactive, got %s", jobType)
	}
}

func TestJobIsNotInteractive(t *testing.T) {
	job := getRunaiJob()

	objects := []runtime.Object{job}
	kubeClient, runaiclient := getClientWithObject(objects)
	trainer := RunaiTrainer{runaijobClient: runaiclient, client: kubeClient.GetClientset()}

	jobs, _ := trainer.ListTrainingJobs(NAMESPACE)

	jobType := jobs[0].Trainer()
	if jobType != "Train" {
		t.Errorf("Expected job to be train, got %s", jobType)
	}
}

func testResourceIncluded(resources []cmdTypes.Resource, name string, resourceType cmdTypes.ResourceType) bool {
	for _, resource := range resources {
		if resource.ResourceType == resourceType && resource.Name == name {
			return true
		}
	}
	return false
}

func createPodOwnedBy(podName string, labelSelector map[string]string, ownerUUID string, ownerKind string, ownerName string) *v1.Pod {
	controller := true
	
	if labelSelector == nil {
		labelSelector = map[string]string{}
	}
	// adding the created by runai label
	labelSelector[createdByLabelKey] = createdByRunAILabelValue

	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: NAMESPACE,
			Labels:    labelSelector,
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					UID:        types.UID(ownerUUID),
					Kind:       ownerKind,
					Name:       ownerName,
					Controller: &controller,
				},
			}},
		Spec: v1.PodSpec{
			SchedulerName: constants.SchedulerName,
		},
	}
}
