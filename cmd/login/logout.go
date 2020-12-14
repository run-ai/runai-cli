package login

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/auth/config"
	"github.com/run-ai/runai-cli/pkg/auth/util"
	"github.com/spf13/cobra"
)

func NewLogoutCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:          "logout",
		Short:        "Removes Tokens from the auth config.",
		SilenceUsage: true,
		Args: func(c *cobra.Command, args []string) error {
			if err := cobra.NoArgs(c, args); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeConfig, err := util.ReadKubeConfig()
			if err != nil {
				return fmt.Errorf("failed to parse kubeconfig: %v", err)
			}
			userAuth, ok := kubeConfig.AuthInfos[paramKubeConfigUser]
			if !ok {
				return fmt.Errorf("No auth configuration found in kubeconfig for user '%s' \n", paramKubeConfigUser)
			}
			if userAuth.AuthProvider != nil && len(userAuth.AuthProvider.Config) > 0 {
				delete(userAuth.AuthProvider.Config, config.IdToken)
				delete(userAuth.AuthProvider.Config, config.RefreshToken)
			} else {
				return fmt.Errorf("No auth configuration found in kubeconfig for user '%s' \n", paramKubeConfigUser)
			}
			if err := util.WriteKubeConfig(kubeConfig); err != nil {
				return err
			}
			fmt.Printf("Auth tokens for user '%s' have been removed.\n", paramKubeConfigUser)
			return nil
		},
	}
	command.Flags().StringVar(&paramKubeConfigUser, "kube-config-user", config.DefaultKubeConfigUserName, "The user defined in the kubeconfig file to operate on, by default a user called runai-oidc is used")

	return command
}
