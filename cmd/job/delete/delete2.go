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

package cmd

import (
	"fmt"

	"github.com/run-ai/runai-cli/cmd/trainer"
	"github.com/run-ai/runai-cli/pkg/client"
)

type JobIdentifier struct {
	name string
	namespace string
	trainer string
	isInteractive bool
}

func generateConflictMessage(conflictedJobs []trainer.TrainingJob) string {
	message := fmt.Sprintln("Conflicted jobs: ")
	for i, job := range conflictedJobs {
		message = fmt.Sprintf("%s \t %d) %s, %s\n", message, i, job.Name(), job.)
	}
	return message
}

func guessTrainingJob(job JobIdentifier, kubeClient *client.Client) (trainer.TrainingJob, error) {
	var matchingJobs []trainer.TrainingJob
	trainers := trainer.NewTrainers(kubeClient)
	for _, trainer := range trainers {
		trainingJob, err := trainer.GetTrainingJob(job.name, job.namespace)
		if err != nil {
			continue
		}
		matchingJobs = append(matchingJobs, trainingJob)
	}

	if len(matchingJobs) == 0 {
		return nil, fmt.Errorf("there is not job name %s in namespace %s", job.name, job.namespace)
	} else if len(matchingJobs) > 1 {
		return nil, fmt.Errorf(generateConflictMessage(matchingJobs))
	}
	return matchingJobs[0], nil
}

func getTrainingJob(job JobIdentifier, kubeClient *client.Client) (trainer.TrainingJob, error) {
	var jobTrainer trainer.Trainer
	var trainingJob trainer.TrainingJob

	var err error
	switch job.trainer {
	case "runaijob":
		jobTrainer = trainer.NewRunaiTrainer(*kubeClient)
	case "mpijob":
		jobTrainer = trainer.NewMPIJobTrainer(*kubeClient)
	default:
		trainingJob, err = guessTrainingJob(job, kubeClient)
		if err != nil {
			return nil, err
		}
		return trainingJob, nil
	}

	trainingJob, err = jobTrainer.GetTrainingJob(job.name, job.namespace)
	fmt.Println(trainingJob.Trainer())
	if err != nil {
		return nil, err
	}

	return trainingJob, nil
}

func DeleteJob(maybeJobIdentifier JobIdentifier, kubeClient *client.Client) error {
	trainingJob, err := getTrainingJob(maybeJobIdentifier, kubeClient)
	if err != nil {
		return err
	}

	fmt.Print("Want to delete")
	fmt.Println(trainingJob)
	err = nil //workflow.DeleteJob(validatedJob, kubeClient)
	if err != nil {
		return err
	}

	return nil
}
