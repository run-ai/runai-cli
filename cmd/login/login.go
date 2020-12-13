package login

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/auth"
	"github.com/run-ai/runai-cli/pkg/auth/jwt"
	"github.com/run-ai/runai-cli/pkg/auth/oidc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	"strings"
)

const (
	KubeConfigUserName                = "runai-oidc"
	AuthMethodBrowser                 = "browser"
	AuthMethodRemoteBrowser           = "remote-browser"
	AuthMethodPassword                = "password"
	AuthMethodLocalClusterIdpPassword = "local-cluster-password" //auth0 and keycloak handle 'password' grant types a bit differently. This flag is for keycloak. The difference is mainly in the redirect url.

	DefaultListenAddress = "127.0.0.1:8000"
	DefaultIssuerUrl     = "https://runai-prod.auth0.com/"
	DefaultRedirectUrl   = "https://app.run.ai/auth"
)

var (
	paramForce          bool
	paramKubeConfigUser string
	paramAuthMethod     string
	paramListenAddress  string
	paramClientId       string
	paramClientSecret   string
	paramIssuerUrl      string
	paramRedirectUrl    string
)

func NewLoginCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:          "login",
		Short:        "It logs you in", // TODO [by dan]:
		SilenceUsage: true,
		Args: func(c *cobra.Command, args []string) error {
			if err := cobra.NoArgs(c, args); err != nil {
				return err
			}
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
			authProviderConfig, err := getOrCreateAuthProviderConfig(kubeConfig)
			authenticator, err := oidc.NewAuthenticator(authProviderConfig)
			if err != nil {
				return err
			}
			var tokens *oidc.KubectlTokens
			var authError error
			switch paramAuthMethod {
			case AuthMethodBrowser:
				options := oidc.BrowserAuthOptions{
					ListenAddress: paramListenAddress,
					ExtraParams:   make(map[string]string), // TODO [by dan]: pass from flag
				}
				tokens, authError = authenticator.BrowserAuth(options)
			case AuthMethodRemoteBrowser:
				tokens, authError = authenticator.RemoteBrowserAuth()

				//case AuthMethodPassword:

				//case AuthMethodLocalClusterIdpPassword:

			}
			if authError != nil {
				return err
			}
			newAuthProviderConfig := authenticator.ToAuthProviderConfig(tokens)
			kubeConfig.AuthInfos[paramKubeConfigUser].AuthProvider = &newAuthProviderConfig
			err = oidc.WriteKubeConfig(kubeConfig)
			if err == nil {
				log.Info("Configuration has been updated. You have logged in successfully.")
			} else {
				err = fmt.Errorf("failed to save configuration with new auth info: %w", err)
			}
			return err
		},
	}

	command.Flags().BoolVar(&paramForce, "force", false, "Force re-authentication even if a valid config was found")
	command.Flags().StringVar(&paramKubeConfigUser, "kube-config-user", KubeConfigUserName, "The user defined in the kubeconfig file to operate on, by default a user called runai-oidc is used")
	command.Flags().StringVar(&paramAuthMethod, "authentication-method", AuthMethodBrowser, "The method to use for initial authentication. can be one of [browser,remote-browser,password]")
	command.Flags().StringVar(&paramListenAddress, "listen-address", DefaultListenAddress, "[browser only] Address to bind to the local server that accepts redirected auth responses.")

	command.Flags().StringVar(&paramClientId, "client-id", "", "OIDC Client ID")
	command.Flags().StringVar(&paramClientSecret, "client-secret", "", "OIDC Client Secret")
	command.Flags().StringVar(&paramIssuerUrl, "issuer-url", DefaultIssuerUrl, "OIDC Issuer URL")
	command.Flags().StringVar(&paramRedirectUrl, "redirect-url", DefaultRedirectUrl, "Auth Response Redirect URL")

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

	// TODO [by dan]: look at oauth2.Token struct which does this already
	if userAuth.AuthProvider != nil && len(userAuth.AuthProvider.Config) > 0 && jwt.IsTokenValid(userAuth.AuthProvider.Config[oidc.IdToken]) {
		return false
	}
	return true
}

func getOrCreateAuthProviderConfig(kubeConfig *clientapi.Config) (authProviderConfig oidc.AuthProviderConfig, err error) {
	shouldPrompt := true
	// If there's an existing config try to validate and use it for login
	if userAuth, ok := kubeConfig.AuthInfos[paramKubeConfigUser]; ok {
		if userAuth.AuthProvider != nil {
			log.Infof("Found an existing Authentication Provider Config for user %s, attempting to use it to re-login.", paramKubeConfigUser)
			if authProviderConfig, err = oidc.ProviderConfig(userAuth.AuthProvider); err != nil {
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
	return authProviderConfig, err
}

// promptUserForAuthProvider prompts user for basic required config so that we can query the identity provider for a token.
func promptUserForAuthProvider() (config oidc.AuthProviderConfig, err error) {
	config.AuthMethod = paramAuthMethod
	if paramClientId == "" {
		if clientId, err := auth.ReadString("Client ID: "); err == nil {
			config.ClientId = clientId
		} else {
			return config, err
		}
	} else {
		config.ClientId = paramClientId
	}

	if paramClientSecret == "" {
		if clientSecret, err := auth.ReadPassword("Client Secret: "); err == nil {
			config.ClientSecret = clientSecret
		} else {
			return config, err
		}
	} else {
		config.ClientSecret = paramClientSecret
	}

	if paramIssuerUrl == "" {
		if issuerUrl, err := auth.ReadString("Issuer URL: "); err == nil {
			config.IssuerUrl = issuerUrl
		} else {
			return config, err
		}
	} else {
		config.IssuerUrl = paramIssuerUrl
	}
	// Make sure issuer url ends with '/'
	issuerUrl := strings.TrimSuffix(config.IssuerUrl, "/")
	config.IssuerUrl = issuerUrl + "/"

	// Use 'Remote Browser' when a browser is not locally available for the local session (i.e while using runai cli via ssh)
	// In this case the cli cannot open a browser and cannot listen for the auth response on localhost:8000 (default) so the redirect must bounce the browser to some generally
	// available location like app.run.ai/auth or <airgapped-backencd-url>/auth for airgapped envs.
	if config.AuthMethod == AuthMethodRemoteBrowser {
		config.RedirectUrl = paramRedirectUrl
	}
	return
}
