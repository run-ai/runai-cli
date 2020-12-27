package kubeconfig

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	ClientIdFieldName  = "client-id"
	IssuerUrlFieldName = "issuer-url"
)

func GetCurrentUserAuthenticationParams() (*types.AuthenticationParams, error) {
	kubeconfig, err := ReadKubeConfig()
	if err != nil {
		return nil, err
	}
	kubeconfigCurrentUser := kubeconfig.AuthInfos[kubeconfig.Contexts[kubeconfig.CurrentContext].AuthInfo]

	clientId, exists := kubeconfigCurrentUser.AuthProvider.Config[ClientIdFieldName]
	if !exists {
		return nil, fmt.Errorf("%v field must be supllied in the kubeconfig file", ClientIdFieldName)
	}
	issuerUrl, exists := kubeconfigCurrentUser.AuthProvider.Config[IssuerUrlFieldName]
	if !exists {
		return nil, fmt.Errorf("%v field must be supllied in the kubeconfig file", IssuerUrlFieldName)
	}
	return &types.AuthenticationParams{
		ClientId:           clientId,
		IssuerURL:          issuerUrl,
		AuthenticationFlow: types.CodePkceBrowser,
	}, nil
	return &types.AuthenticationParams{
		ClientId:           "86fdLjb3G0A5pNtKpE3H6vFDWis9sL6I",
		IssuerURL:          "https://runai-test.auth0.com/",
		ListenAddress:      "localhost:3000",
		AuthenticationFlow: types.CodePkceBrowser,
	}, nil
}

func ReadKubeConfig() (*api.Config, error) {
	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	return configAccess.GetStartingConfig()
}

func WriteKubeConfig(config *api.Config) error {
	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	return clientcmd.ModifyConfig(configAccess, *config, true)
}
