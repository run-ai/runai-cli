// +build test

package trainer

import (
	clientset "github.com/run-ai/runai-cli/cmd/mpi/client/clientset/versioned"
	"github.com/run-ai/runai-cli/pkg/client"
)

// Create a RunaiTrainer with given clients. for test purposes.
func NewRunaiTrainerWithClients(client client.Client, runaiclient clientset.Interface) Trainer {
	return &RunaiTrainer{
		client:         client.GetClientset(),
		runaijobClient: runaiclient,
	}
}
