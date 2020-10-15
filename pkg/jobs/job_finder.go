package jobs

import (
	"fmt"
	"github.com/run-ai/runai-cli/cmd/trainer"
	"github.com/run-ai/runai-cli/pkg/client"
)

const (InteractiveJobTrainerLabel = "Interactive"
		TrainJobTrainerLabel = "Train"
		PreemptibleInteractiveJobTrainerLabel = "Preemptible-Interactive")

type JobIdentifier struct {
	Name          string
	Namespace     string
	Trainer       string
	Interactive   bool
	Train		  bool
}

func generateConflictError(conflictedJobs []trainer.TrainingJob) error {
	message := fmt.Sprintf("There are more than one training job with the name %s: \n", conflictedJobs[0].Name())
	for i, job := range conflictedJobs {
		message = fmt.Sprintf("%s \t %d) %s, %s, %s\n", message, i, job.Name(), job.TrainerName(), job.Type())
	}
	message = fmt.Sprintf("%sTo delete a specifig job you can use the flags --training-type, --interactive, and --train", message)
	return fmt.Errorf(message)
}

func guessTrainingJobByTrainer(job JobIdentifier, t trainer.Trainer) ([]trainer.TrainingJob, error) {
	trainingJobs, err := t.GetTrainingJobs(job.Name, job.Namespace)
	if err != nil {
		return nil, err
	}

	var matchingJobs []trainer.TrainingJob
	for _, trainingJob := range trainingJobs {
		if !job.Interactive && !job.Train{
			matchingJobs = append(matchingJobs, trainingJob)
		} else if job.Interactive && (trainingJob.Type() == InteractiveJobTrainerLabel || trainingJob.Type() == PreemptibleInteractiveJobTrainerLabel) {
			matchingJobs = append(matchingJobs, trainingJob)
		} else if job.Train && (trainingJob.Type() == TrainJobTrainerLabel) {
			matchingJobs = append(matchingJobs, trainingJob)
		}
	}
	return matchingJobs, nil
}

func guessTrainingJob(job JobIdentifier, kubeClient *client.Client) (trainer.TrainingJob, error) {
	var matchingJobs []trainer.TrainingJob
	trainers := trainer.NewTrainers(kubeClient)
	for _, trainer := range trainers {
		trainingJobs, err := guessTrainingJobByTrainer(job, trainer)
		if err != nil {
			continue
		}
		matchingJobs = append(matchingJobs, trainingJobs...)
	}

	if len(matchingJobs) == 0 {
		return nil, fmt.Errorf("there is not job name %s in namespace %s", job.Name, job.Namespace)
	} else if len(matchingJobs) > 1 {
		return nil, generateConflictError(matchingJobs)
	}
	return matchingJobs[0], nil
}

func getTrainingJobByTrainer(job JobIdentifier, t trainer.Trainer) (trainer.TrainingJob, error){
	trainingJobs, err := t.GetTrainingJobs(job.Name, job.Namespace)
	if err != nil {
		return nil, err
	}

	if len(trainingJobs) == 1 {
		return trainingJobs[0], nil
	} else if len(trainingJobs) > 1 {
		if !job.Interactive && !job.Train {
			return nil, generateConflictError(trainingJobs)
		}
		for _, trainingJob := range trainingJobs {
			if job.Interactive && (trainingJob.Type() == InteractiveJobTrainerLabel || trainingJob.Type() == PreemptibleInteractiveJobTrainerLabel) {
				return trainingJob, nil
			} else if job.Train && (trainingJob.Type() == TrainJobTrainerLabel) {
				return trainingJob, nil
			}
		}
		return nil, generateConflictError(trainingJobs)
	}
	return nil, fmt.Errorf("could not find a job with name: %s, trainer: %s, interactive: %t, namespace: %s", job.Name, job.Trainer, job.Interactive, job.Namespace)
}

func GetTrainingJob(job JobIdentifier, kubeClient *client.Client) (trainer.TrainingJob, error) {
	var jobTrainer trainer.Trainer
	var trainingJob trainer.TrainingJob

	var err error
	switch job.Trainer {
	case trainer.RunaiJobTrainerName:
		jobTrainer = trainer.NewRunaiTrainer(*kubeClient)
	case trainer.MpiJobTrainerName:
		jobTrainer = trainer.NewMPIJobTrainer(*kubeClient)
	default:
		trainingJob, err = guessTrainingJob(job, kubeClient)
		if err != nil {
			return nil, err
		}
		return trainingJob, nil
	}

	return getTrainingJobByTrainer(job, jobTrainer)
}
