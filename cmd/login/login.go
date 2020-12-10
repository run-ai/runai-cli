package login

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/auth/jwt"
	"github.com/run-ai/runai-cli/pkg/auth/oidc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	KubeConfigUserName                = "runai-oidc"
	AuthMethodBrowser                 = "browser"
	AuthMethodRemoteBrowser           = "remote-browser"
	AuthMethodPassword                = "password"
	AuthMethodLocalClusterIdpPassword = "local-cluster-password" //auth0 and keycloak handle 'password' grant types a bit differently. This flag is for keycloak
	DefaultListenAddress			  = "127.0.0.1:8000"
)

var (
	force          bool
	kubeConfigUser string
	authMethod     string
	listenAddress  string
)

func NewLoginCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "login",
		Short: "It logs you in", // TODO [by dan]:
		Args: func(c *cobra.Command, args []string) error {
			if err := cobra.NoArgs(c, args); err != nil {
				return err
			}
			//if o.IssuerURL == "" {
			//	return errors.New("--oidc-issuer-url is missing")
			//}
			//if o.ClientID == "" {
			//	return errors.New("--oidc-client-id is missing")
			//}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			kubeConfig, err := oidc.ReadKubeConfig()
			if err != nil {
				return fmt.Errorf("failed to parse kubeconfig: %v", err)
			}

			if !shouldDoAuth(kubeConfig) {
				// Can still override this with --force
				log.Info("Current configuration seems valid, no need to login again.")
				return nil
			}

			authProviderConfig, err := getOrCreateAuthProviderConfig(kubeConfig, err)

			authenticator, err := oidc.NewAuthenticator(authProviderConfig)
			if err != nil {
				return err
			}

			var tokens *oidc.KubectlTokens
			switch authMethod {
			case AuthMethodBrowser:
				options := oidc.BrowserAuthOptions{
					ListenAddress: listenAddress,
					ExtraParams:   make(map[string]string), // TODO [by dan]: pass from flag
				}
				tokens, err = authenticator.BrowserAuth(options)
				if err != nil {
					return err
				}
				//case AuthMethodRemoteBrowser:

				//case AuthMethodPassword:

				//case AuthMethodLocalClusterIdpPassword:

			}

			newAuthProviderConfig := authenticator.ToAuthProviderConfig(tokens)
			kubeConfig.AuthInfos[kubeConfigUser].AuthProvider = &newAuthProviderConfig
			err = oidc.WriteKubeConfig(kubeConfig)
			if err == nil {
				log.Info("Configuration has been updated. You have logged in successfully.")
			} else {
				err = fmt.Errorf("failed to save configuration with new auth info: %w", err)
			}
			return err
		},
	}

	command.Flags().BoolVar(&force, "force", false, "Force re-authentication even if a valid config was found")
	command.Flags().StringVar(&kubeConfigUser, "kube-config-user", KubeConfigUserName, "The user defined in the kubeconfig file to operate on, by default a user called runai-oidc is used")
	command.Flags().StringVar(&authMethod, "authentication-method", AuthMethodBrowser, "The method to use for initial authentication. can be one of [browser,remote-browser,password]")
	command.Flags().StringVar(&listenAddress, "listen-address", DefaultListenAddress, "[browser only] Address to bind to the local server that accepts redirected auth responses.")

	return command
}

func getOrCreateAuthProviderConfig(kubeConfig *clientapi.Config, err error) (oidc.AuthProviderConfig, error) {
	var authProviderConfig oidc.AuthProviderConfig
	shouldPrompt := true
	// If there's an existing config try to validate and use it for login
	if userAuth, ok := kubeConfig.AuthInfos[kubeConfigUser]; ok {
		if userAuth.AuthProvider != nil {
			log.Infof("Found an existing Authentication Provider Config for user %s, attempting to use it to re-login.", kubeConfigUser)
			if authProviderConfig, err = oidc.ProviderConfig(userAuth.AuthProvider); err != nil {
				log.Warnf("An auth provider config exists for user %s but is invalid: %v", kubeConfigUser, err)
			} else {
				shouldPrompt = false
			}
		}
	} else {
		// No user found in config, create it
		kubeConfig.AuthInfos[kubeConfigUser] = &clientapi.AuthInfo{}
	}
	if shouldPrompt {
		authProviderConfig = promptUserForAuthProvider()
	}
	return authProviderConfig, err
}

func shouldDoAuth(kubeConfig *clientapi.Config) bool {
	if force {
		return true
	}

	userAuth, ok := kubeConfig.AuthInfos[kubeConfigUser]
	if !ok {
		log.Infof("No auth configuration found in kubeconfig for user '%s' ", kubeConfigUser)
		return true
	}

	// TODO [by dan]: look at oauth2.Token struct which does this already
	if userAuth.AuthProvider != nil && len(userAuth.AuthProvider.Config) > 0 && jwt.IsTokenValid(userAuth.AuthProvider.Config[oidc.IdToken]) {
		return false
	}

	return true

}

func promptUserForAuthProvider() (config oidc.AuthProviderConfig) {
	// TODO [by dan]:

	return
}