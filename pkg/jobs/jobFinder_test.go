package jobs

import (
	"github.com/magiconair/properties/assert"
	"github.com/run-ai/runai-cli/cmd/trainer"
	cmdTypes "github.com/run-ai/runai-cli/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

type trainerMock struct {
	trainingJobs []trainer.TrainingJob
}

func (t trainerMock) IsSupported(name, ns string) bool { return true }
func (t trainerMock) Type() string { return "Mock" }
func (t trainerMock) IsEnabled() bool { return true }
func (t trainerMock) ListTrainingJobs(namespace string) ([]trainer.TrainingJob, error) {return []trainer.TrainingJob{}, nil}

func (t trainerMock) GetTrainingJobs(name, namespace string) ([]trainer.TrainingJob, error) { return t.trainingJobs, nil}

func TestEmptyNamespace(t *testing.T) {
	job := JobIdentifier{Name: "test1", Namespace: "runaitest", Trainer: "mpijob", Train: false, Interactive: false}
	trainer := trainerMock{}

	guessedJobs, err := guessTrainingJobByTrainer(job, trainer)
	if err != nil {
		t.Errorf("guessTrainingJobByTrainer failed in case of empty namespace test: %v", err)
	}
	assert.Equal(t, len(guessedJobs), 0)
}

func TestSingleRunaiNoFlag(t *testing.T) {
	job := JobIdentifier{Name: "test1", Namespace: "runaitest", Trainer: "runaijob", Train: false, Interactive: false}
	trainerJob := newRunaiJob("test1", "runaitest", true)
	trainer := trainerMock{trainingJobs: []trainer.TrainingJob{trainerJob}}

	guessedJobs, err := guessTrainingJobByTrainer(job, trainer)
	if err != nil {
		t.Errorf("Error threw in the time of the test: %v", err)
	}
	assert.Equal(t, len(guessedJobs), 1)
}

func TestSingleRunaiInteractive(t *testing.T) {
	job := JobIdentifier{Name: "test1", Namespace: "runaitest", Trainer: "runaijob", Train: false, Interactive: true}
	trainerJob := newRunaiJob("test1", "runaitest", true)
	trainer := trainerMock{trainingJobs: []trainer.TrainingJob{trainerJob}}

	guessedJobs, err := guessTrainingJobByTrainer(job, trainer)
	if err != nil {
		t.Errorf("Error threw in the time of the test: %v", err)
	}
	assert.Equal(t, len(guessedJobs), 1)
}

func TestSingleRunaiTrain(t *testing.T) {
	job := JobIdentifier{Name: "test1", Namespace: "runaitest", Trainer: "runaijob", Train: true, Interactive: false}
	trainerJob := newRunaiJob("test1", "runaitest", false)
	trainer := trainerMock{trainingJobs: []trainer.TrainingJob{trainerJob}}

	guessedJobs, err := guessTrainingJobByTrainer(job, trainer)
	if err != nil {
		t.Errorf("Error threw in the time of the test: %v", err)
	}
	assert.Equal(t, len(guessedJobs), 1)
}

func TestMultipleRunai(t *testing.T) {
	job := JobIdentifier{Name: "test1", Namespace: "runaitest", Trainer: "runaijob", Train: false, Interactive: false}
	trainerJobs := []trainer.TrainingJob{}
	trainerJobs = append(trainerJobs, newRunaiJob("test1", "runaitest", true))
	trainerJobs = append(trainerJobs, newRunaiJob("test1", "runaitest", false))

	trainer := trainerMock{trainingJobs: trainerJobs}

	guessedJobs, err := guessTrainingJobByTrainer(job, trainer)
	if err != nil {
		t.Errorf("Error threw in the time of the test: %v", err)
	}
	assert.Equal(t, len(guessedJobs), 2)
}

func newRunaiJob(name, namespace string, interactive bool) trainer.TrainingJob {
	trainingType := TrainJobTrainerLabel
	if interactive {
		trainingType = InteractiveJobTrainerLabel
	}
	return trainer.NewRunaiJob(nil,
		nil,
		metav1.Time{},
		trainingType,
		name,
		true,
		nil,
		false,
		v1.PodSpec{},
		metav1.ObjectMeta{},
		metav1.ObjectMeta{},
		namespace,
		cmdTypes.Resource{},
		"",
		0,
		0,
		0,
		0)
}