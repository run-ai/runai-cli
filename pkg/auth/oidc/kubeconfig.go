package oidc

import (
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	ClientId             = "client-id"
	ClientSecret         = "client-secret"
	IdToken              = "id-token"
	RefreshToken         = "refresh-token"
	IssuerUrl            = "idp-issuer-url"
	RedirectUrl			 = "redirect-url"
)

func ReadKubeConfig() (*clientcmdapi.Config, error) {
	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	return configAccess.GetStartingConfig()
}

func WriteKubeConfig(config *clientcmdapi.Config) error {
	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	return clientcmd.ModifyConfig(configAccess, *config, true)
}

func ProviderConfig(authConfig *clientcmdapi.AuthProviderConfig) (config AuthProviderConfig, err error) {
	if clientId, exists := authConfig.Config[ClientId]; !exists {
		return config, fmt.Errorf("missing field in Auth Provider Config: %s", ClientId)
	} else {
		config.ClientId = clientId
	}

	if clientSecret, exists := authConfig.Config[ClientSecret]; !exists {
		return config, fmt.Errorf("missing field in Auth Provider Config: %s", ClientSecret)
	} else {
		config.ClientSecret = clientSecret
	}

	if issuerUrl, exists := authConfig.Config[IssuerUrl]; !exists {
		return config, fmt.Errorf("missing field in Auth Provider Config: %s", issuerUrl)
	} else {
		config.IssuerUrl = issuerUrl
	}

	if redirectUrl, exists := authConfig.Config[RedirectUrl]; exists {
		config.RedirectUrl = redirectUrl
	}

	if idToken, exists := authConfig.Config[IdToken]; exists {
		config.IdToken = idToken
	}

	if refreshToken, exists := authConfig.Config[RefreshToken]; exists {
		config.RefreshToken = refreshToken
	}

	return config, nil
}
