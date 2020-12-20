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
	"strings"
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
	paramAuthRealm      string
	paramListenAddress  string
	paramClientId       string
	paramClientSecret   string
	paramIssuerUrl      string
	paramRedirectUrl    string
	paramExtraScopes    string
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
			if !shouldDoAuth(kubeConfig) {
				// Can still override this with --force
				log.Info("Current configuration seems valid, no need to login again.")
				return nil
			}
			authProviderConfig, err := getOrCreateAuthProviderConfig(kubeConfig)
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

	// Required
	command.Flags().StringVar(&paramClientId, ParamClientId, "", "OIDC Client ID")
	command.Flags().StringVar(&paramClientSecret, ParamClientSecret, "", "OIDC Client Secret")

	// Required, but has defaults
	command.Flags().StringVar(&paramIssuerUrl, ParamIssuerUrl, DefaultIssuerUrl, "OIDC Issuer URL")
	command.Flags().StringVar(&paramRedirectUrl, ParamRedirectUrl, DefaultRedirectUrl, "Auth Response Redirect URL")
	command.Flags().StringVar(&paramAuthMethod, ParamAuthMethod, DefaultAuthMethod, "The method to use for initial authentication. can be one of [browser,remote-browser,password,local-password]")
	command.Flags().StringVar(&paramKubeConfigUser, "kube-config-user", DefaultKubeConfigUserName, "The user defined in the kubeconfig file to operate on, by default a user called runai-oidc is used")
	command.Flags().StringVar(&paramListenAddress, ParamListenAddress, DefaultListenAddress, "[browser only] Address to bind to the local server that accepts redirected auth responses.")

	// Optional
	command.Flags().StringVar(&paramAuthRealm, ParamAuthRealm, "", "[password only] Governs which realm will be used when authenticating the user with the IDP")
	command.Flags().StringVar(&paramExtraScopes, ParamExtraScopes, "", "Extra scopes to request with the ID token.")

	return command
}

func shouldDoAuth(kubeConfig *clientapi.Config) bool {
	if paramForce {
		return true
	}
	userAuth, ok := kubeConfig.AuthInfos[paramKubeConfigUser]
	if !ok {
		fmt.Printf("No auth configuration found in kubeconfig for user '%s' \n", paramKubeConfigUser)
		return true
	}

	if userAuth.AuthProvider != nil && len(userAuth.AuthProvider.Config) > 0 && jwt.IsTokenValid(userAuth.AuthProvider.Config[ParamIdToken]) {
		return false
	}
	return true
}

func getOrCreateAuthProviderConfig(kubeConfig *clientapi.Config) (authProviderConfig AuthProviderConfig, err error) {
	shouldPrompt := true
	// If there's an existing config try to validate and use it for login
	if userAuth, ok := kubeConfig.AuthInfos[paramKubeConfigUser]; ok {
		if userAuth.AuthProvider != nil {
			fmt.Printf("Found an existing Authentication Provider Config for user %s, attempting to use it to re-login.\n", paramKubeConfigUser)
			if authProviderConfig, err = ProviderConfig(userAuth.AuthProvider); err != nil {
				log.Warnf("An auth provider config exists for user %s but is invalid: %v", paramKubeConfigUser, err)
			} else {
				shouldPrompt = false
			}
		}
	} else {
		// No user found in config, create it
		kubeConfig.AuthInfos[paramKubeConfigUser] = &clientapi.AuthInfo{}
	}
	if shouldPrompt {
		authProviderConfig, err = promptUserForAuthProvider()
	}
	assignParamsIfNeeded(&authProviderConfig)
	return authProviderConfig, err
}

// promptUserForAuthProvider prompts user for basic required config so that we can query the identity provider for a token.
func promptUserForAuthProvider() (config AuthProviderConfig, err error) {
	if paramClientId == "" {
		if clientId, err := util.ReadString("Client ID: "); err == nil {
			config.ClientId = clientId
		} else {
			return config, err
		}
	} else {
		config.ClientId = paramClientId
	}

	if paramClientSecret == "" {
		if clientSecret, err := util.ReadPassword("Client Secret: "); err == nil {
			config.ClientSecret = clientSecret
		} else {
			return config, err
		}
	} else {
		config.ClientSecret = paramClientSecret
	}
	return
}

// param vars hold defaults so this actually setting user overrides and/or defaults where needed.
func assignParamsIfNeeded(config *AuthProviderConfig) {
	if config.AuthMethod == DefaultAuthMethod || config.AuthMethod == "" {
		config.AuthMethod = paramAuthMethod
	}
	if config.ListenAddress == DefaultListenAddress || config.ListenAddress == "" {
		config.ListenAddress = paramListenAddress
	}
	if config.IssuerUrl == DefaultIssuerUrl || config.IssuerUrl == "" {
		config.IssuerUrl = paramIssuerUrl
	}
	if config.RedirectUrl == DefaultRedirectUrl || config.RedirectUrl == "" {
		config.RedirectUrl = paramRedirectUrl
	}
	if config.AuthRealm == "" {
		config.AuthRealm = paramAuthRealm
	}
	if paramExtraScopes != "" {
		config.ExtraScopes = strings.Split(paramExtraScopes, ExtraScopesSeparator)
	}
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
