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
		Short: "Authenticates your client with the Run:AI Backend",
		Run: func(cmd *cobra.Command, args []string) {
			log.Debugf("starting authentication [cli args: %v, authentication params cli: %v]", args, params)
			err := authentication.Authenticate(params)
			if err != nil {
				cmd.HelpFunc()(cmd, args)
				log.Error(err)
				os.Exit(1)
			}
			log.Info("Logged in successfully")
		},
	}
	command.Flags().StringVar(&params.ClientId, "client-id", "", "Client id to connect")
	command.Flags().StringVar(&params.IssuerURL, "idp-issuer-url", "", "issuer url")
	command.Flags().StringVar(&params.ListenAddress, "redirect-server", "", "listen address")
	command.Flags().StringVar(&params.User, "user", "", "user to log in")
	command.Flags().MarkHidden("client-id")
	command.Flags().MarkHidden("idp-issuer-url")
	command.Flags().MarkHidden("redirect-server")
	command.Flags().MarkHidden("user")

	return command
}
