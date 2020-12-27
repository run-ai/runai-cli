package logout

import (
	"github.com/run-ai/runai-cli/pkg/authentication/logout"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func NewLogoutCommand() *cobra.Command {
	var user string
	var command = &cobra.Command{
		Use:   "logout",
		Short: "Logout from runai",
		Run: func(cmd *cobra.Command, args []string) {
			log.Debugf("Logout user. cli args: %v, cli user param: %v", args, user)
			err := logout.Logout(user)
			if err != nil {
				cmd.HelpFunc()(cmd, args)
				log.Error(err)
				os.Exit(1)
			}
			log.Info("Logged out successfully")
		},
	}
	command.Flags().StringVar(&user, "user", "", "user to log out")

	return command
}
