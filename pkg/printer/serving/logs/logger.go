package logs

import (
	servejob "github.com/run-ai/runai-cli/pkg/jobs/serving"
	"github.com/run-ai/runai-cli/pkg/podlogs"
	podlogger "github.com/run-ai/runai-cli/pkg/printer/base/logs"
)

type ServingPodLogPrinter struct {
	PodLogger *podlogger.PodLogPrinter
}

func NewServingPodLogPrinter(job servejob.Serving, logArgs *podlogs.OuterRequestArgs) (*ServingPodLogPrinter, error) {
	var names []string
	for _, pod := range job.AllPods() {
		names = append(names, pod.ObjectMeta.Name)
	}

	podLogPrinter, err := podlogger.NewPodLogPrinter(names, logArgs)
	if err != nil {
		return nil, err
	}
	return &ServingPodLogPrinter{
		PodLogger: podLogPrinter,
	}, nil
}

func (slp *ServingPodLogPrinter) Print() (int, error) {
	return slp.PodLogger.Print()
}
