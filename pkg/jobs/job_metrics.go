package jobs

import (
	"strconv"
	"strings"

	"github.com/prometheus/common/log"
	"github.com/run-ai/runai-cli/cmd/trainer"
	prom "github.com/run-ai/runai-cli/pkg/prometheus"
	"github.com/run-ai/runai-cli/pkg/types"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// prometheus query names
	gpuAllocationPQ   = "gpuAllocation"
	gpuUtilizationPQ  = "gpuUtilization"
	usedGpusMemoryPQ  = "usedGpusMemory"
	usedCpusMemoryPQ  = "usedCpusMemory"
	utilizedCpusPQ    = "utilizedCpus"
	usedCpusPQ        = "usedCpus"
	requestedCpusPQ   = "requestedCpus"
	requestedCPUMemPQ = "requestedCPUMem"
)

type metricSetter struct {
	Name      string
	QueryName string
	Setter    func(*types.JobView, prom.MetricValue)
}

var metricSetters = []metricSetter{
	{
		Name:      "GPU Allocation",
		QueryName: gpuAllocationPQ,
		Setter: func(job *types.JobView, value prom.MetricValue) {
			job.GPUs.Allocated = value.(float64)
		},
	},
	{
		Name:      "GPU Utilization",
		QueryName: gpuUtilizationPQ,
		Setter: func(job *types.JobView, value prom.MetricValue) {
			job.GPUs.Utilization = value.(float64)
		},
	},
	{
		Name:      "GPU Memory",
		QueryName: usedGpusMemoryPQ,
		Setter: func(job *types.JobView, value prom.MetricValue) {
			job.GPUMem.Usage.Usage = value.(float64) * trainer.GpuMbFactor
			if job.GPUMem.Allocated != 0 {
				job.GPUMem.Usage.Utilization = (job.GPUMem.Usage.Usage / job.GPUMem.Allocated) * 100
			}
		},
	},
	{
		Name:      "CPU Requested",
		QueryName: requestedCpusPQ,
		Setter: func(job *types.JobView, value prom.MetricValue) {
			job.CPUs.Allocated = value.(float64)
		},
	},
	{
		Name:      "CPU Utilization",
		QueryName: utilizedCpusPQ,
		Setter: func(job *types.JobView, value prom.MetricValue) {
			if job.CPUs.Usage == nil {
				job.CPUs.Usage = &types.ResourceUsage{}
			}
			job.CPUs.Usage.Utilization = value.(float64)
		},
	},
	{
		Name:      "CPU Usage",
		QueryName: usedCpusPQ,
		Setter: func(job *types.JobView, value prom.MetricValue) {
			if job.CPUs.Usage == nil {
				job.CPUs.Usage = &types.ResourceUsage{}
			}
			job.CPUs.Usage.Usage = value.(float64)
		},
	},
	{
		Name:      "CPU Memory Requested",
		QueryName: requestedCPUMemPQ,
		Setter: func(job *types.JobView, value prom.MetricValue) {
			job.Mem.Allocated = value.(float64)
		},
	},
	{
		Name:      "CPU Memory Usage",
		QueryName: usedCpusMemoryPQ,
		Setter: func(job *types.JobView, value prom.MetricValue) {
			job.Mem.Usage.Usage = value.(float64)
			if job.Mem.Allocated != 0 {
				job.Mem.Usage.Utilization = (job.Mem.Usage.Usage / job.Mem.Allocated) * 100
			}
		},
	},
}

var (
	prometheusJobLabelID = "pod_group_uuid"
	jobPQs               = prom.QueryNameToQuery{
		gpuAllocationPQ:   `runai_allocated_gpus`,
		gpuUtilizationPQ:  `sum(runai_pod_group_gpu_utilization) by (pod_group_uuid) / on (pod_group_uuid) (count(runai_pod_group_gpu_utilization) by (pod_group_uuid))`,
		usedGpusMemoryPQ:  `sum(runai_pod_group_used_gpu_memory) by (pod_group_uuid)`,
		usedCpusMemoryPQ:  `runai_job_memory_used_bytes`,
		requestedCpusPQ:   `runai_active_job_cpu_allocated_cores`,
		utilizedCpusPQ:    `runai_job_cpu_usage / on(pod_group_uuid) runai_active_job_cpu_allocated_cores * 100`,
		usedCpusPQ:        `runai_job_cpu_usage`,
		requestedCPUMemPQ: `runai_active_job_memory_allocated_bytes`,
	}
)

func getJobAllocatedGPUMem(job trainer.TrainingJob) float64 {
	memoryQuantity, err := resource.ParseQuantity(job.CurrentAllocatedGPUsMemory())
	if err != nil {
		log.Warn("Couldn't read allocated memory for job ", job.Name(), " : ", err)
		memoryQuantity = *resource.NewQuantity(0, resource.DecimalSI)
	}
	allocatedMemory, ok := memoryQuantity.AsInt64()
	if !ok {
		return 0
	}
	return float64(allocatedMemory)
}

func trainingJobToJobView(jobs []trainer.TrainingJob) map[string]types.JobView {
	views := make(map[string]types.JobView, len(jobs))
	for _, jobInfo := range jobs {
		podGroupUUID := jobInfo.GetPodGroupUUID()
		nodeName := jobInfo.HostIPOfChief()
		if strings.Contains(nodeName, ", ") {
			nodeName = "<multiple>"
		}
		views[podGroupUUID] = types.JobView{
			Info: &types.JobGeneralInfo{
				Name:     jobInfo.Name(),
				Project:  jobInfo.Project(),
				User:     jobInfo.User(),
				Type:     jobInfo.Trainer(),
				Status:   jobInfo.GetStatus(),
				Duration: jobInfo.Duration(),
				Node:     nodeName,
			},
			GPUs: &types.GPUMetrics{
				Allocated: jobInfo.CurrentAllocatedGPUs(),
			},
			GPUMem: &types.MemoryMetrics{
				Allocated: getJobAllocatedGPUMem(jobInfo),
				Usage:     &types.ResourceUsage{},
			},
			CPUs: &types.CPUMetrics{
				Usage: &types.ResourceUsage{},
			},
			Mem: &types.MemoryMetrics{
				Usage: &types.ResourceUsage{},
			},
		}
	}
	return views
}

func queryJobsMetrics(promClient prom.QueryClient) (*prom.MetricResultsByItems, error) {
	var promData prom.MetricResultsByItems
	promData, promErr := promClient.GroupMultiQueriesToItems(jobPQs, prometheusJobLabelID)
	if promErr != nil {
		return nil, promErr
	}
	return &promData, nil
}

func addMetricsDataToViews(jobs map[string]types.JobView, metrics prom.MetricResultsByItems) {
	for podGroupUUID, job := range jobs {
		jobMetrics, found := metrics[podGroupUUID]
		if !found {
			log.Debugln("Couldn't find metrics for job: ", job.Info.Name, " pod group uuid: ", podGroupUUID)
			continue
		}
		for _, metricSetterInfo := range metricSetters {
			metricResult, found := jobMetrics[metricSetterInfo.QueryName]
			if found {
				if metricValue, err := strconv.ParseFloat((*metricResult)[0].Value[1].(string), 64); err != nil {
					log.Debugln("Failed to convert ", metricSetterInfo.Name, ": ", err)
				} else {
					metricSetterInfo.Setter(&job, metricValue)
				}
			} else {
				log.Debug("Metric ", metricSetterInfo.Name, " is missing for job ", job.Info.Name)
			}
		}
	}
}

// GetJobsMetrics fetches and returns information about all requested jobs
func GetJobsMetrics(client prom.QueryClient, jobs []trainer.TrainingJob) (views []types.JobView, err error) {
	jobsInfo := trainingJobToJobView(jobs)
	metrics, err := queryJobsMetrics(client)
	if err == nil {
		addMetricsDataToViews(jobsInfo, *metrics)
	}

	views = make([]types.JobView, 0, len(jobs))
	for _, job := range jobs {
		views = append(views, jobsInfo[job.GetPodGroupUUID()])
	}

	return views, err
}
