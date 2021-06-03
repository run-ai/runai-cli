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
	"sort"

	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/types"
	v1 "k8s.io/api/core/v1"
)

const (
	DefaultRunaiTrainingType = "runai"
)

// construct the trainer list
func NewTrainers(kubeClient *client.Client) []Trainer {

	trainers := []Trainer{}
	trainerInits := []func(kubeClient client.Client) Trainer{
		// NewHorovodJobTrainer,
		// NewStandaloneJobTrainer,
		// NewTensorFlowJobTrainer,
		NewMPIJobTrainer,
		// NewSparkJobTrainer,
		// NewVolcanoJobTrainer,
		NewRunaiTrainer}

	for _, init := range trainerInits {
		trainers = append(trainers, init(*kubeClient))
	}

	return trainers
}

type orderedTrainingJob []TrainingJob

func (this orderedTrainingJob) Len() int {
	return len(this)
}

func (this orderedTrainingJob) Less(i, j int) bool {
	return this[i].RequestedGPU() > this[j].RequestedGPU()
}

func (this orderedTrainingJob) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

type orderedTrainingJobByAge []TrainingJob

func (this orderedTrainingJobByAge) Len() int {
	return len(this)
}

func (this orderedTrainingJobByAge) Less(i, j int) bool {
	if this[i].StartTime() == nil {
		return true
	} else if this[j].StartTime() == nil {
		return false
	}

	return this[i].StartTime().After(this[j].StartTime().Time)
}

func (this orderedTrainingJobByAge) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

type orderedTrainingJobByName []TrainingJob

func (job orderedTrainingJobByName) Len() int {
	return len(job)
}

func (job orderedTrainingJobByName) Less(i, j int) bool {
	return job[i].Name() < job[j].Name()
}

func (job orderedTrainingJobByName) Swap(i, j int) {
	job[i], job[j] = job[j], job[i]
}

type orderedTrainingJobByProject []TrainingJob

func (job orderedTrainingJobByProject) Len() int {
	return len(job)
}

func (job orderedTrainingJobByProject) Less(i, j int) bool {
	return job[i].Project() < job[j].Project()
}

func (job orderedTrainingJobByProject) Swap(i, j int) {
	job[i], job[j] = job[j], job[i]
}

func MakeTrainingJobOrderdByAge(jobList []TrainingJob) []TrainingJob {
	newJoblist := make(orderedTrainingJobByAge, 0, len(jobList))
	for _, v := range jobList {
		newJoblist = append(newJoblist, v)
	}
	sort.Sort(newJoblist)
	return []TrainingJob(newJoblist)
}

func MakeTrainingJobOrderdByName(jobList []TrainingJob) []TrainingJob {
	var newJoblist orderedTrainingJobByName
	for _, job := range jobList {
		newJoblist = append(newJoblist, job)
	}
	sort.Sort(newJoblist)
	return newJoblist
}

func MakeTrainingJobOrderdByProject(jobList []TrainingJob) []TrainingJob {
	var newJoblist orderedTrainingJobByProject
	for _, job := range jobList {
		newJoblist = append(newJoblist, job)
	}
	sort.Sort(newJoblist)
	return newJoblist
}

func MakeTrainingJobOrderdByGPUCount(jobList []TrainingJob) []TrainingJob {
	newJoblist := make(orderedTrainingJob, 0, len(jobList))
	for _, v := range jobList {
		newJoblist = append(newJoblist, v)
	}
	sort.Sort(newJoblist)
	return []TrainingJob(newJoblist)
}

func contains(s []v1.PodPhase, searchterm string) bool {
	for _, a := range s {
		if a == v1.PodPhase(searchterm) {
			return true
		}
	}
	return false
}

// GetAllJobs and filter them by `namespaceInfo` and optionaly filters out pod `filterStatus`
func GetAllJobs(kubeClient *client.Client, namespaceInfo types.NamespaceInfo, filterStatus []v1.PodPhase) (jobs []TrainingJob, err error) {
	trainers := NewTrainers(kubeClient)
	for _, trainer := range trainers {
		if !trainer.IsEnabled() {
			continue
		}
		trainingJobs, err := trainer.ListTrainingJobs(namespaceInfo.Namespace)
		if err != nil {
			return nil, err
		}
		if len(filterStatus) != 0 {
			for _, job := range trainingJobs {
				if filterStatus != nil && contains(filterStatus, job.GetStatus()) {
					jobs = append(jobs, job)
				}
			}
		} else {
			jobs = append(jobs, trainingJobs...)
		}
	}
	return jobs, nil
}
