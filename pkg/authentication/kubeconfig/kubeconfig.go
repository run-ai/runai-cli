package kubeconfig

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	"golang.org/x/oauth2"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	clientIdFieldName           = "client-id"
	issuerUrlFieldName          = "idp-issuer-url"
	idTokenFieldName            = "id-token"
	refreshTokenFieldName       = "refresh-token"
	idTokenRawTokenName         = "id_token"
	authenticationFlowFieldName = "auth-flow"
)

func GetCurrentUserAuthenticationParams() (*types.AuthenticationParams, error) {
	kubeConfig, err := readKubeConfig()
	if err != nil {
		return nil, err
	}
	return getUserAuthenticationParams(kubeConfig.Contexts[kubeConfig.CurrentContext].AuthInfo, kubeConfig)
}

func GetUserAuthenticationParams(user string) (*types.AuthenticationParams, error) {
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

func getUserAuthenticationParams(user string, kubeConfig *api.Config) (*types.AuthenticationParams, error) {
	kubeConfigUser, exists := kubeConfig.AuthInfos[user]
	if !exists {
		return nil, fmt.Errorf("user %v does not exists in kubeconfig", user)
	}
	if kubeConfigUser.AuthProvider == nil {
		return &types.AuthenticationParams{}, nil
	}

	clientId, exists := kubeConfigUser.AuthProvider.Config[clientIdFieldName]
	if !exists {
		return nil, fmt.Errorf("%v field must be supllied in the kubeConfig file", clientIdFieldName)
	}
	issuerUrl, exists := kubeConfigUser.AuthProvider.Config[issuerUrlFieldName]
	if !exists {
		return nil, fmt.Errorf("%v field must be supllied in the kubeConfig file", issuerUrlFieldName)
	}
	authenticationFlow := kubeConfigUser.AuthProvider.Config[authenticationFlowFieldName]
	return &types.AuthenticationParams{
		ClientId:           clientId,
		IssuerURL:          issuerUrl,
		AuthenticationFlow: authenticationFlow,
	}, nil
}

func setTokenToUser(user, authenticationFlow string, token *oauth2.Token, kubeConfig *api.Config) error {
	kubeConfigUser, exists := kubeConfig.AuthInfos[user]
	if !exists {
		return fmt.Errorf("user %v does not exists in kubeconfig", user)
	}
	if idToken := token.Extra(idTokenRawTokenName); idToken != nil {
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
