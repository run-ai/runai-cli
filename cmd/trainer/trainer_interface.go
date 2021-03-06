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

package trainer

import (
	"time"

	cmdTypes "github.com/run-ai/runai-cli/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// The Training Job can be TensorFlow, MPI and Caffe
type TrainingJob interface {
	// Get the chief Pod of the Job.
	ChiefPod() *v1.Pod

	// Get the name of the Training Job
	Name() string

	// Get the namespace of the Training Job
	Namespace() string

	// Get all the pods of the Training Job
	AllPods() []v1.Pod

	// Get all the kubernetes resource of the Training Job
	Resources() []cmdTypes.Resource

	// Get the Status of the Job: RUNNING, PENDING,
	GetStatus() string

	// Return trainer Type, support MPI, standalone, tensorflow
	Trainer() string

	// Get the Job Age
	Age() time.Duration

	// Get the Job Duration
	Duration() time.Duration

	// Get start time
	StartTime() *metav1.Time

	// Get Dashboard
	GetJobDashboards(client *kubernetes.Clientset) ([]string, error)

	// Requested GPU count of the Job
	RequestedGPU() float64

	RequestedGPUMemory() uint64

	RequestedGPUString() string

	// Requested GPU count of the Job
	AllocatedGPU() float64

	// the host ip of the chief pod
	HostIPOfChief() string

	// The priority class name of the training job
	GetPriorityClass() string

	Project() string

	User() string

	Image() string

	CreatedByCLI() bool

	ServiceURLs() []string

	GetPodGroupName() string
	GetPodGroupUUID() string

	RunningPods() int32
	PendingPods() int32
	Completions() int32
	Parallelism() int32
	Succeeded() int32
	Failed() int32
	CurrentRequestedGPUs() float64
	CurrentRequestedGPUsMemory() int64
	CurrentAllocatedGPUs() float64
	CurrentAllocatedGPUsMemory() string
	WorkloadType() string
	TotalRequestedGPUsString() string
	CurrentRequestedGpusString() string
	CliCommand() string
}

// Trainer interface for querying specific types of training jobs
type Trainer interface {
	// Check if the training job is supported
	IsSupported(name, ns string) bool

	// Get TrainingJob object directly. this method is called when `arena get`
	GetTrainingJob(name, namespace string) (TrainingJob, error)

	// Get the type of trainer
	Type() string

	ListTrainingJobs(namespace string) ([]TrainingJob, error)

	// Returns whether the trainer is enabled
	IsEnabled() bool
}
