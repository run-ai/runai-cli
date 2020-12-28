package kubeconfig

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/authentication/authentication-params"
	"golang.org/x/oauth2"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	IdTokenRawTokenName         = "id_token"
	clientIdFieldName           = "client-id"
	issuerUrlFieldName          = "idp-issuer-url"
	idTokenFieldName            = "id-token"
	refreshTokenFieldName       = "refresh-token"
	authenticationFlowFieldName = "auth-flow"
	auth0RealmFieldName         = "auth0-realm"
)

func GetCurrentUserAuthenticationParams() (*authentication_params.AuthenticationParams, error) {
	kubeConfig, err := readKubeConfig()
	if err != nil {
		return nil, err
	}
	return getUserAuthenticationParams(kubeConfig.Contexts[kubeConfig.CurrentContext].AuthInfo, kubeConfig)
}

func GetUserAuthenticationParams(user string) (*authentication_params.AuthenticationParams, error) {
	kubeConfig, err := readKubeConfig()
	if err != nil {
		return nil, err
	}
	return getUserAuthenticationParams(user, kubeConfig)
}

func SetTokenToCurrentUser(authenticationFlow string, token *oauth2.Token) error {
	kubeConfig, err := readKubeConfig()
	if err != nil {
		return err
	}
	return setTokenToUser(kubeConfig.Contexts[kubeConfig.CurrentContext].AuthInfo, authenticationFlow, token, kubeConfig)
}

func SetTokenToUser(user, authenticationFlow string, token *oauth2.Token) error {
	kubeConfig, err := readKubeConfig()
	if err != nil {
		return err
	}
	return setTokenToUser(user, authenticationFlow, token, kubeConfig)
}

func DeleteTokenToCurrentUser() error {
	kubeConfig, err := readKubeConfig()
	if err != nil {
		return err
	}
	return deleteTokenToUser(kubeConfig.Contexts[kubeConfig.CurrentContext].AuthInfo, kubeConfig)
}

func DeleteTokenToUser(user string) error {
	kubeConfig, err := readKubeConfig()
	if err != nil {
		return err
	}
	return deleteTokenToUser(user, kubeConfig)
}

func getUserAuthenticationParams(user string, kubeConfig *api.Config) (*authentication_params.AuthenticationParams, error) {
	kubeConfigUser, exists := kubeConfig.AuthInfos[user]
	if !exists {
		return nil, fmt.Errorf("user %v does not exists in kubeconfig", user)
	}
	if kubeConfigUser.AuthProvider == nil {
		return &authentication_params.AuthenticationParams{}, nil
	}

	clientId := kubeConfigUser.AuthProvider.Config[clientIdFieldName]
	issuerUrl := kubeConfigUser.AuthProvider.Config[issuerUrlFieldName]
	authenticationFlow := kubeConfigUser.AuthProvider.Config[authenticationFlowFieldName]
	auth0Realm := kubeConfigUser.AuthProvider.Config[auth0RealmFieldName]

	return &authentication_params.AuthenticationParams{
		ClientId:           clientId,
		IssuerURL:          issuerUrl,
		AuthenticationFlow: authenticationFlow,
		Auth0Realm:         auth0Realm,
	}, nil
}

func setTokenToUser(user, authenticationFlow string, token *oauth2.Token, kubeConfig *api.Config) error {
	kubeConfigUser, exists := kubeConfig.AuthInfos[user]
	if !exists {
		return fmt.Errorf("user %v does not exists in kubeconfig", user)
	}
	if idToken := token.Extra(IdTokenRawTokenName); idToken != nil {
		kubeConfigUser.AuthProvider.Config[idTokenFieldName] = idToken.(string)
	}
	kubeConfigUser.AuthProvider.Config[refreshTokenFieldName] = token.RefreshToken
	kubeConfigUser.AuthProvider.Config[authenticationFlowFieldName] = authenticationFlow

	return writeKubeConfig(kubeConfig)
}

func deleteTokenToUser(user string, kubeConfig *api.Config) error {
	kubeConfigUser, exists := kubeConfig.AuthInfos[user]
	if !exists {
		return fmt.Errorf("user %v does not exists in kubeconfig", user)
	}
	if _, exists := kubeConfigUser.AuthProvider.Config[idTokenFieldName]; exists {
		delete(kubeConfigUser.AuthProvider.Config, idTokenFieldName)
	}
	if _, exists := kubeConfigUser.AuthProvider.Config[refreshTokenFieldName]; exists {
		delete(kubeConfigUser.AuthProvider.Config, refreshTokenFieldName)
	}

	return writeKubeConfig(kubeConfig)
}

func readKubeConfig() (*api.Config, error) {
	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	return configAccess.GetStartingConfig()
}

func writeKubeConfig(config *api.Config) error {
	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	return clientcmd.ModifyConfig(configAccess, *config, true)
}
