package login

import (
	"github.com/run-ai/runai-cli/pkg/authentication"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func NewLoginCommand() *cobra.Command {
	params := &types.AuthenticationParams{}
	var command = &cobra.Command{
		Use:   "login",
		Short: "Login to runai",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := authentication.Authenticate(params)
			if err != nil {
				cmd.HelpFunc()(cmd, args)
				log.Error(err)
				os.Exit(1)
			}
		},
	}
	command.Flags().StringVar(&params.ClientId, "client-id", "", "Client id to connect")
	command.Flags().StringVar(&params.IssuerURL, "issuer-url", "", "issuer url")
	command.Flags().StringVar(&params.ListenAddress, "redirect-server", "", "listen address")

	return command
}
