package login

import (
	"fmt"
	. "github.com/run-ai/runai-cli/pkg/auth/config"
	"github.com/run-ai/runai-cli/pkg/auth/jwt"
	"github.com/run-ai/runai-cli/pkg/auth/oidc"
	"github.com/run-ai/runai-cli/pkg/auth/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	// Use 'Remote Browser' when a browser is not locally available for the local session (i.e while using runai cli via ssh)
	// In this case the cli cannot open a browser and cannot listen for the auth response on localhost:8000 (default) so the redirect must bounce the browser to some generally
	// available location like app.run.ai/auth or <airgapped-backencd-url>/auth for airgapped envs.
	AuthMethodRemoteBrowser           = "remote-browser"
	AuthMethodBrowser                 = "browser"
	AuthMethodPassword                = "password"
	AuthMethodLocalClusterIdpPassword = "local-cluster-password" //auth0 and keycloak handle 'password' grant types a bit differently. This flag is for keycloak. The difference is mainly in the redirect url.
)

var (
	paramForce          bool
	paramKubeConfigUser string
	paramAuthMethod     string
	paramListenAddress  string
)

func NewLoginCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:          "login",
		Short:        "Authenticates your client with the Run:AI Backend",
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
			userAuth, err := getUserAuth(kubeConfig)
			if err != nil {
				return err
			}

			if !shouldDoAuth(userAuth) {
				// Can still override this with --force
				log.Info("Current configuration seems valid, no need to login again.")
				return nil
			}
			authProviderConfig, err := getAuthProviderConfig(kubeConfig)
			if err != nil {
				return err
			}
			authenticator, err := oidc.NewAuthenticator(authProviderConfig)
			if err != nil {
				return err
			}
			var tokens *Tokens
			switch authProviderConfig.AuthMethod {
			case AuthMethodBrowser:
				tokens, err = authenticator.BrowserAuth(paramListenAddress)
			case AuthMethodRemoteBrowser:
				tokens, err = authenticator.RemoteBrowserAuth()
			case AuthMethodPassword:
				tokens, err = authenticator.Auth0PasswordAuth()
			case AuthMethodLocalClusterIdpPassword:
				tokens, err = authenticator.PasswordAuth()
			default:
				err = fmt.Errorf("unknown auth method: %s", authProviderConfig.AuthMethod)
			}
			if err != nil {
				return err
			}
			return writeAuthProviderConfigToKubeConfig(authProviderConfig, tokens, kubeConfig)
		},
	}

	command.Flags().BoolVar(&paramForce, "force", false, "Force re-authentication even if a valid config was found")
	command.Flags().StringVar(&paramAuthMethod, ParamAuthMethod, DefaultAuthMethod, "The method to use for initial authentication. can be one of [browser,remote-browser,password,local-password]")
	command.Flags().StringVar(&paramKubeConfigUser, "kube-config-user", DefaultKubeConfigUserName, "The user defined in the kubeconfig file to operate on, by default a user called runai-oidc is used")
	command.Flags().StringVar(&paramListenAddress, ParamListenAddress, DefaultListenAddress, "[browser only] Address to bind to the local server that accepts redirected auth responses.")

	return command
}

func shouldDoAuth(userAuth *clientapi.AuthInfo) bool {
	if paramForce {
		return true
	}
	if userAuth.AuthProvider != nil && len(userAuth.AuthProvider.Config) > 0 && jwt.IsTokenValid(userAuth.AuthProvider.Config[ParamIdToken]) {
		return false
	}
	return true
}

func getUserAuth(kubeConfig *clientapi.Config) (userAuth *clientapi.AuthInfo, err error) {
	ok := false
	currentUser := kubeConfig.Contexts[kubeConfig.CurrentContext].AuthInfo
	userAuth, ok = kubeConfig.AuthInfos[paramKubeConfigUser]
	if !ok {
		// Try to look for auth info in current context user
		userAuth, ok = kubeConfig.AuthInfos[currentUser]
		if ok {
			paramKubeConfigUser = currentUser
		}
	}
	if !ok {
		err = fmt.Errorf("No auth configuration found in kubeconfig for user '%s' \n", paramKubeConfigUser)
	}
	return userAuth, err
}

func getAuthProviderConfig(kubeConfig *clientapi.Config) (authProviderConfig AuthProviderConfig, err error) {
	// If there's an existing config try to validate and use it for login
	if userAuth, ok := kubeConfig.AuthInfos[paramKubeConfigUser]; ok {
		if userAuth.AuthProvider != nil {
			fmt.Printf("Found an existing Authentication Provider Config for user %s, attempting to use it to re-login.\n", paramKubeConfigUser)
			if authProviderConfig, err = ProviderConfig(userAuth.AuthProvider); err != nil {
				return AuthProviderConfig{}, fmt.Errorf("an auth provider config exists for user %s but is invalid: %v", paramKubeConfigUser, err)
			}
		}
	} else {
		// No user found in config, fail.
		return authProviderConfig, fmt.Errorf("no auth provider config found for user %s", paramKubeConfigUser)
	}
	err = assignAndValidateParams(&authProviderConfig)
	return authProviderConfig, err
}

// param vars hold defaults so this actually setting user overrides and/or defaults where needed.
func assignAndValidateParams(config *AuthProviderConfig) error {
	if config.AuthMethod == DefaultAuthMethod || config.AuthMethod == "" {
		config.AuthMethod = paramAuthMethod
	}
	if config.ListenAddress == DefaultListenAddress || config.ListenAddress == "" {
		config.ListenAddress = paramListenAddress
	}
	if config.IssuerUrl == "" {
		return fmt.Errorf("issuer URL must be defined")
	}
	if config.RedirectUrl == "" {
		return fmt.Errorf("redirect URL must be defined")
	}
	return nil
}

func writeAuthProviderConfigToKubeConfig(authProviderConfig AuthProviderConfig, tokens *Tokens, kubeConfig *clientapi.Config) error {
	authProviderConfig.AddTokens(tokens)
	newAuthProviderConfig := authProviderConfig.ToKubeAuthProviderConfig()
	kubeConfig.AuthInfos[paramKubeConfigUser].AuthProvider = &newAuthProviderConfig
	err := util.WriteKubeConfig(kubeConfig)
	if err == nil {
		fmt.Println("Configuration has been updated. You have logged in successfully.")
	} else {
		err = fmt.Errorf("failed to save configuration with new auth info: %v", err)
	}
	return err
}
