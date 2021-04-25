// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
)

const (
	RunaiGPUIndex    = "runai-gpu"
	RunaiGPUFraction = "gpu-fraction"
	RunaiGPUMemory   = "gpu-memory"
	// an annotation on each node
	AllocatableGpus                    = "nvidia.com/gpu.count"
	NVIDIAGPUResourceName              = "nvidia.com/gpu"
	ALIYUNGPUResourceName              = "aliyun.com/gpu-mem"
	DeprecatedNVIDIAGPUResourceName    = "alpha.kubernetes.io/nvidia-gpu"
	PodGroupRequestedGPUs              = "runai-podgroup-requested-gpus"
	PodGroupRequestedGPUsMemory        = "runai-podgroup-requested-gpus-memory"
	CliCommand        				   = "runai-cli-command"
	WorkloadCurrentAllocatedGPUs       = "runai-current-allocated-gpus"
	WorkloadCurrentAllocatedGPUsMemory = "runai-current-allocated-gpus-memory"
	WorkloadCurrentRequestedGPUs       = "runai-current-requested-gpus"
	WorkloadCurrentRequestedGPUsMemory = "runai-current-requested-gpus-memory"
	WorkloadTotalRequestedGPUs         = "runai-total-requested-gpus"
	WorkloadTotalRequestedGPUsMemory   = "runai-total-requested-gpus-memory"
)

// The way to get total GPU Count of Node: nvidia.com/gpu
func TotalGpuInNode(node v1.Node) int64 {
	val, ok := node.Status.Capacity[NVIDIAGPUResourceName]

	if !ok {
		return GpuInNodeDeprecated(node)
	}

	return val.Value()
}

// The way to get allocatble GPU Count of Node
func AllocatableGpuInNodeIncludingFractions(node v1.Node) int64 {
	val, ok := node.Labels[AllocatableGpus]

	if ok {
		gpus, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			return gpus
		}
	}

	return TotalGpuInNode(node)
}

// The way to get GPU Count of Node: alpha.kubernetes.io/nvidia-gpu
func GpuInNodeDeprecated(node v1.Node) int64 {
	val, ok := node.Status.Allocatable[DeprecatedNVIDIAGPUResourceName]

	if !ok {
		return 0
	}

	return val.Value()
}

func GpuInPod(pod v1.Pod) (gpuCount int64) {
	containers := pod.Spec.Containers
	for _, container := range containers {
		gpuCount += gpuInContainer(container)
	}

	return gpuCount
}

func GpuInActivePod(pod v1.Pod) (gpuCount float64) {
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return 0
	}

	gpuFractionUsed := getGPUFractionUsedByPod(pod)
	if gpuFractionUsed > 0 {
		return gpuFractionUsed
	}

	return float64(GpuInPod(pod))
}

func GetRequestedGPUsPerPodGroup(trainingAnnotations map[string]string) (float64, bool) {
	if len(trainingAnnotations[PodGroupRequestedGPUs]) > 0 {
		requestedGPUs, err := strconv.ParseFloat(trainingAnnotations[PodGroupRequestedGPUs], 64)
		if err == nil {
			return requestedGPUs, true
		}
	}
	return 0, false
}

func GetRequestedGPUsMemoryPerPodGroup(trainingAnnotations map[string]string) uint64 {
	if len(trainingAnnotations[PodGroupRequestedGPUsMemory]) > 0 {
		requestedGPUs, err := strconv.ParseUint(trainingAnnotations[PodGroupRequestedGPUsMemory], 10, 64)
		if err == nil {
			return requestedGPUs
		}
	}
	return 0
}

func GetRequestedGPUString(trainingAnnotations map[string]string) string {
	if len(trainingAnnotations[PodGroupRequestedGPUs]) > 0 {
		requestedGPUs, err := strconv.ParseFloat(trainingAnnotations[PodGroupRequestedGPUs], 64)
		if err == nil {
			return fmt.Sprintf("%v", requestedGPUs)
		}
	}
	return fmt.Sprintf("%v", 0)
}

func getGPUFractionUsedByPod(pod v1.Pod) float64 {
	if pod.Annotations != nil {
		gpuFraction, GPUFractionErr := strconv.ParseFloat(pod.Annotations[RunaiGPUFraction], 64)
		if GPUFractionErr == nil {
			return gpuFraction
		}
	}

	return 0
}

func gpuInContainer(container v1.Container) int64 {
	val, ok := container.Resources.Limits[NVIDIAGPUResourceName]

	if !ok {
		return GpuInContainerDeprecated(container)
	}

	return val.Value()
}

func GpuInContainerDeprecated(container v1.Container) int64 {
	val, ok := container.Resources.Limits[DeprecatedNVIDIAGPUResourceName]

	if !ok {
		return 0
	}

	return val.Value()
}

func GetSharedGPUsIndexUsedInPods(pods []v1.Pod) map[string]float64 {
	gpuIndexUsed := map[string]float64{}
	for _, pod := range pods {
		if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
			continue
		}

		if pod.Annotations != nil {
			gpuIndex, found := pod.Annotations[RunaiGPUIndex]
			if !found {
				continue
			}

			gpuAllocated := getGPUFractionUsedByPod(pod)

			gpuIndexUsed[gpuIndex] += gpuAllocated
		}
	}

	return gpuIndexUsed
}
