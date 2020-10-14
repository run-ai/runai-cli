package jobs

import (
	"fmt"
	"github.com/run-ai/runai-cli/cmd/trainer"
	"github.com/run-ai/runai-cli/pkg/client"
)

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
		message = fmt.Sprintf("%s \t %d) %s, %s, %s\n", message, i, job.Name(), job.TrainerName(), job.Trainer())
	}
	message = fmt.Sprintf("%sTo delete a specifig job you can use the flags --training-type, --interactive, and --train", message)
	return fmt.Errorf(message)
}

func guessTrainingJob(job JobIdentifier, kubeClient *client.Client) (trainer.TrainingJob, error) {
	var matchingJobs []trainer.TrainingJob
	trainers := trainer.NewTrainers(kubeClient)
	for _, trainer := range trainers {
		trainingJobs, err := trainer.GetTrainingJobs(job.Name, job.Namespace)
		if err != nil {
			continue
		}

		if job.Interactive {
			for _, trainingJob := range trainingJobs {
				if (trainingJob.Trainer() == "Interactive") == job.Interactive {
					matchingJobs = append(matchingJobs, trainingJob)
				}
			}
		} else if job.Train{
			for _, trainingJob := range trainingJobs {
				if (trainingJob.Trainer() == "Train") == job.Train {
					matchingJobs = append(matchingJobs, trainingJob)
				}
			}
		} else {
			matchingJobs = append(matchingJobs, trainingJobs...)
		}
	}

	if len(matchingJobs) == 0 {
		return nil, fmt.Errorf("there is not job name %s in namespace %s", job.Name, job.Namespace)
	} else if len(matchingJobs) > 1 {
		return nil, generateConflictError(matchingJobs)
	}
	return matchingJobs[0], nil
}

func GetTrainingJob(job JobIdentifier, kubeClient *client.Client) (trainer.TrainingJob, error) {
	var jobTrainer trainer.Trainer
	var trainingJob trainer.TrainingJob

	var err error
	switch job.Trainer {
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

	trainingJobs, err := jobTrainer.GetTrainingJobs(job.Name, job.Namespace)
	if err != nil {
		return nil, err
	}

	if len(trainingJobs) == 1 {
		return trainingJobs[0], nil
	} else if job.Interactive {
		for _, trainingJob := range trainingJobs {
			if (trainingJob.Trainer() == "Interactive") == job.Interactive {
				return trainingJob, nil
			}
		}
	} else if job.Train{
		for _, trainingJob := range trainingJobs {
			if (trainingJob.Trainer() == "Train") == job.Train {
				return trainingJob, nil
			}
		}
	} else if len(trainingJobs) > 1 {
		return nil, generateConflictError(trainingJobs)
	}
	return nil, fmt.Errorf("could not find a job with name: %s, trainer: %s, interactive: %s, namespace: %s", job.Name, job.Trainer, job.Interactive, job.Namespace)
}
