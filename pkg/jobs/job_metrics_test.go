package jobs

import (
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/square/go-jose.v2/json"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/run-ai/runai-cli/cmd/trainer"
	cmdutil "github.com/run-ai/runai-cli/cmd/util"
	prom "github.com/run-ai/runai-cli/pkg/prometheus"
	"github.com/run-ai/runai-cli/pkg/types"
	cmdTypes "github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/util"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	NAMESPACE = "namespace"
)

var ()

func ReadMetricsFromFile(filename string) (metrics prom.MetricResultsByItems, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &metrics)
	return
}

func TestJobInfo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Job information and metrics collection test")
}

var _ = Describe("Job Information Collection", func() {
	var (
		expectedJobView types.JobView
	)
	BeforeEach(func() {
		expectedJobView = types.JobView{
			Info: &types.JobGeneralInfo{
				Name:     "job-name",
				Project:  "test_project",
				User:     "test_user",
				Type:     "Train",
				Status:   "Pending",
				Duration: 0,
				Node:     "test_node",
			},
			GPUs: &types.GPUMetrics{
				Allocated:   1,
				Utilization: 99,
			},
			GPUMem: &types.MemoryMetrics{
				Allocated: 10000 * 1000 * 1000,
				Usage: &types.ResourceUsage{
					Usage:       9000 * 1000 * 1000,
					Utilization: 90,
				},
			},
			CPUs: &types.CPUMetrics{
				Allocated: 0.5,
				Usage: &types.ResourceUsage{
					Usage:       0.75,
					Utilization: 150}},
			Mem: &types.MemoryMetrics{
				Allocated: 104857600,
				Usage: &types.ResourceUsage{
					Usage:       2621440000,
					Utilization: 2500,
				},
			},
		}
	})
	Describe("GetJobsMetrics", func() {
		Context("RunaiJob", func() {
			var (
				job     *batch.Job
				pod     *v1.Pod
				metrics prom.MetricResultsByItems
			)
			BeforeEach(func() {
				var err error
				job = util.GetRunaiJob(NAMESPACE, "job-name", "id1")
				pod = util.CreatePodOwnedBy(NAMESPACE, "pod", nil, string(job.UID), string(cmdTypes.ResourceTypeJob), job.Name)
				metrics, err = ReadMetricsFromFile("example_metrics.json")
				if err != nil {
					Fail(fmt.Sprintf("%v", err))
				}
			})
			It("collects metrics for single job", func() {
				objects := []runtime.Object{pod, job}
				client, runaiclient := util.GetClientWithObject(objects)
				jobTrainer := trainer.NewRunaiTrainerWithClients(client, runaiclient)
				jobs, err := jobTrainer.ListTrainingJobs(NAMESPACE)
				if err != nil {
					Fail(fmt.Sprintf("%v", err))
				}
				fakePromClient := util.FakePrometheusClient(metrics, nil)
				jobViews, _ := GetJobsMetrics(fakePromClient, jobs)
				Expect(jobViews).To(Equal([]types.JobView{expectedJobView}))
			})
			It("Handles missing metrics", func() {
				delete(metrics["id1"], "gpuAllocation")
				expectedJobView.GPUs.Allocated = 0
				delete(metrics["id1"], "requestedCPUMem")
				expectedJobView.Mem.Allocated = 0
				expectedJobView.Mem.Usage.Utilization = 0
				objects := []runtime.Object{pod, job}
				client, runaiclient := util.GetClientWithObject(objects)
				jobTrainer := trainer.NewRunaiTrainerWithClients(client, runaiclient)
				jobs, err := jobTrainer.ListTrainingJobs(NAMESPACE)
				if err != nil {
					Fail(fmt.Sprintf("%v", err))
				}
				fakePromClient := util.FakePrometheusClient(metrics, nil)
				jobViews, _ := GetJobsMetrics(fakePromClient, jobs)
				Expect(jobViews).To(Equal([]types.JobView{expectedJobView}))
			})
			It("Handles no metrics", func() {
				delete(metrics, "id1")
				objects := []runtime.Object{pod, job}
				client, runaiclient := util.GetClientWithObject(objects)
				jobTrainer := trainer.NewRunaiTrainerWithClients(client, runaiclient)
				jobs, err := jobTrainer.ListTrainingJobs(NAMESPACE)
				if err != nil {
					Fail(fmt.Sprintf("%v", err))
				}
				fakePromClient := util.FakePrometheusClient(metrics, nil)
				jobViews, _ := GetJobsMetrics(fakePromClient, jobs)
				expectedJobView.GPUs = &types.GPUMetrics{}
				expectedJobView.CPUs = &types.CPUMetrics{Usage: &types.ResourceUsage{}}
				expectedJobView.Mem = &types.MemoryMetrics{Usage: &types.ResourceUsage{}}
				expectedJobView.GPUMem.Usage = &types.ResourceUsage{}
				Expect(jobViews).To(Equal([]types.JobView{expectedJobView}))
			})
			It("Shows metrics for multiple jobs sorted by the order they were sent in", func() {
				job3 := util.GetRunaiJob(NAMESPACE, "job-3", "id3")
				job3.Annotations[cmdutil.PodGroupRequestedGPUs] = "2"
				pod3 := util.CreatePodOwnedBy(NAMESPACE, "pod3", nil, string(job3.UID), string(cmdTypes.ResourceTypeJob), job3.Name)
				job2 := util.GetRunaiJob(NAMESPACE, "job-2", "id2")
				pod2 := util.CreatePodOwnedBy(NAMESPACE, "pod2", nil, string(job2.UID), string(cmdTypes.ResourceTypeJob), job2.Name)
				objects := []runtime.Object{pod, job, job3, pod2, job2, pod3}
				client, runaiclient := util.GetClientWithObject(objects)
				jobTrainer := trainer.NewRunaiTrainerWithClients(client, runaiclient)
				jobs, err := jobTrainer.ListTrainingJobs(NAMESPACE)
				if err != nil {
					Fail(fmt.Sprintf("%v", err))
				}
				jobs = trainer.MakeTrainingJobOrderdByGPUCount(trainer.MakeTrainingJobOrderdByName(jobs))
				metrics[string(job2.UID)] = metrics[string(job.UID)]
				metrics[string(job3.UID)] = metrics[string(job.UID)]
				fakePromClient := util.FakePrometheusClient(metrics, nil)
				jobViews, _ := GetJobsMetrics(fakePromClient, jobs)
				expectedJobView2 := expectedJobView
				expectedJobView2.Info = &types.JobGeneralInfo{
					Name:     "job-2",
					Project:  "test_project",
					User:     "test_user",
					Type:     "Train",
					Status:   "Pending",
					Duration: 0,
					Node:     "test_node",
				}
				expectedJobView3 := expectedJobView
				expectedJobView3.Info = &types.JobGeneralInfo{
					Name:     "job-3",
					Project:  "test_project",
					User:     "test_user",
					Type:     "Train",
					Status:   "Pending",
					Duration: 0,
					Node:     "test_node",
				}
				// sort by GPU count and then by name
				Expect(jobViews).To(Equal([]types.JobView{expectedJobView3, expectedJobView2, expectedJobView}))
			})
		})
	})
})
