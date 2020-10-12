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
	"strconv"

	v1 "k8s.io/api/core/v1"
)

const (
	RunaiGPUIndex                   = "runai-gpu"
	RunaiGPUFraction                = "gpu-fraction"
	// an annotation on each node
	RunaiAllocatableGpus            = "runai-allocatable-gpus"
	NVIDIAGPUResourceName           = "nvidia.com/gpu"
	ALIYUNGPUResourceName           = "aliyun.com/gpu-mem"
	DeprecatedNVIDIAGPUResourceName = "alpha.kubernetes.io/nvidia-gpu"
	PodGroupRequestedGPUs           = "runai-podgroup-requested-gpus"
	WorkloadCurrentAllocatedGPUs    = "runai-current-allocated-gpus"
	WorkloadCurrentRequestedGPUs    = "runai-current-requested-gpus"
	WorkloadTotalRequestedGPUs      = "runai-total-requested-gpus"

)

// filter out the pods with GPU
func GpuPods(pods []v1.Pod) (podsWithGPU []v1.Pod) {
	for _, pod := range pods {
		if GpuInPod(pod) > 0 {
			podsWithGPU = append(podsWithGPU, pod)
		}
	}
	return podsWithGPU
}

// The way to get total GPU Count of Node: nvidia.com/gpu
func TotalGpuInNode(node v1.Node) int64 {
	val, ok := node.Status.Capacity[NVIDIAGPUResourceName]

	if !ok {
		return GpuInNodeDeprecated(node)
	}

	return val.Value()
}

// The way to get allocatble GPU Count of Node
func AllocatableGpuInNode(node v1.Node) (num int64) {
	val, ok := node.Annotations[RunaiAllocatableGpus]

	if ok {
		num , _ = strconv.ParseInt(val, 10, 64)
	}
	
	return 
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

func GetSharedGPUsIndexUsedInPods(pods []v1.Pod) []string {
	gpuIndexUsed := map[string]bool{}
	for _, pod := range pods {
		if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
			continue
		}

		if pod.Annotations != nil {
			gpuIndex, found := pod.Annotations[RunaiGPUIndex]
			if !found {
				continue
			}

			gpuIndexUsed[gpuIndex] = true
		}
	}

	gpuIndexesArray := []string{}

	for key, _ := range gpuIndexUsed {
		gpuIndexesArray = append(gpuIndexesArray, key)
	}

	return gpuIndexesArray
}
