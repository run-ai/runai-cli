package workflow

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	runaijobv1 "github.com/run-ai/runai-cli/cmd/mpi/api/runaijob/v1"
	mpi "github.com/run-ai/runai-cli/cmd/mpi/api/v1alpha2"
	fakeclientset "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned/fake"
	cmdTypes "github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	Namespace = "namespace"
	JobName   = "job-name"
)

var (
	namespaceInfo = cmdTypes.NamespaceInfo{
		Namespace:   Namespace,
		ProjectName: Namespace,
	}
)

func TestJobSuspend(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Job suspend test")
}

var _ = Describe("Job Suspend and Resume", func() {
	Context("RunaiJob", func() {
		var (
			job *runaijobv1.RunaiJob
			pod *v1.Pod
			// client      kubeclient.Client
			runaiclient *fakeclientset.Clientset
		)
		BeforeEach(func() {
			job = util.GetRunaiJob(Namespace, JobName, "id1")
			pod = util.CreatePodOwnedBy(Namespace, "pod", nil, string(job.UID), string(cmdTypes.ResourceTypeJob), job.Name)
		})
		Describe("Job suspned", func() {
			BeforeEach(func() {
				objects := []runtime.Object{pod, job}
				_, runaiclient = util.GetClientWithObject(objects)
			})
			It("Sets suspend field to true", func() {
				SuspendJob(JobName, namespaceInfo, runaiclient)
				updatedJob, err := runaiclient.RunV1().RunaiJobs(Namespace).Get(JobName, metav1.GetOptions{})
				if err != nil {
					Fail(fmt.Sprintf("%v", err))
				}
				Expect(updatedJob.Spec.Suspend).NotTo(BeNil())
				Expect(*updatedJob.Spec.Suspend).To(BeTrue())
			})
		})
		Describe("Job resume", func() {
			BeforeEach(func() {
				trueVal := true
				job.Spec.Suspend = &trueVal
				objects := []runtime.Object{pod, job}
				_, runaiclient = util.GetClientWithObject(objects)
			})
			It("Sets suspend field to false", func() {
				ResumeJob(JobName, namespaceInfo, runaiclient)
				updatedJob, err := runaiclient.RunV1().RunaiJobs(Namespace).Get(JobName, metav1.GetOptions{})
				if err != nil {
					Fail(fmt.Sprintf("%v", err))
				}
				Expect(updatedJob.Spec.Suspend).NotTo(BeNil())
				Expect(*updatedJob.Spec.Suspend).To(BeFalse())
			})
		})
	})
	Context("Not RunaiJob", func() {
		var (
			job         *mpi.MPIJob
			runaiclient *fakeclientset.Clientset
		)
		BeforeEach(func() {
			job = util.GetMPIJob(Namespace, JobName, "id1")
			objects := []runtime.Object{job}
			_, runaiclient = util.GetClientWithObject(objects)
		})
		It("Fails gracefully", func() {
			err := SuspendJob(JobName, namespaceInfo, runaiclient)
			Expect(err).NotTo(BeNil())
		})
	})
})
