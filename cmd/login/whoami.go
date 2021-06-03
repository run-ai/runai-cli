package login

import (
	"fmt"
	"github.com/run-ai/runai-cli/cmd/completion"
	"github.com/run-ai/runai-cli/pkg/authentication"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func NewWhoamiCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "whoami",
		Short: "Current logged in user",
		ValidArgsFunction: completion.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			subject, email, err := authentication.GetCurrentAuthenticateUserSubject()
			if err != nil {
				if errStr := err.Error(); strings.Contains(errStr, "authProvider.config does not exists") {
					log.Info("You are currently not logged in to Run:AI")
				} else {
					log.Error(err)
				}
				os.Exit(1)
			}
			log.Info(fmt.Sprintf("User: %s\nLogged in Id: %s", email, subject))
		},
	}

	return command
}