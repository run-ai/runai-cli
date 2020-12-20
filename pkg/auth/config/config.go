package config

import (
	"fmt"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	"strings"
)

const (
	ParamClientId        = "client-id"
	ParamClientSecret    = "client-secret"
	ParamIdToken         = "id-token"
	ParamRefreshToken    = "refresh-token"
	ParamIssuerUrl       = "idp-issuer-url"
	ParamRedirectUrl     = "redirect-url"
	ParamAuthMethod      = "auth-method"
	ParamExtraScopes     = "auth-request-extra-scopes"
	ParamAuthRealm       = "auth-realm"
	ParamListenAddress   = "listen-address"

	ExtraScopesSeparator = ","
	AuthProviderName     = "oidc"

	DefaultKubeConfigUserName = "runai-oidc"
	DefaultListenAddress      = "127.0.0.1:8000"
	DefaultAuthMethod         = "browser"
)

type Tokens struct {
	IdToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`

	// These are available in the json if you need them, but OIDC only requires the ID token
	//AccessToken  string `json:"access_token,omitempty"`
	//ExpiresIn    int64  `json:"expires_in,omitempty"`
}

type BrowserAuthOptions struct {
	ListenAddress string
}

// These are the minimal required fields to use kubectl's oidc auth plugin
type AuthProviderConfig struct {
	AuthMethod    string
	AuthRealm     string // Only used with auth-method 'Password' (which is actually the auth0 variant)
	ClientId      string
	ClientSecret  string
	IdToken       string
	RefreshToken  string
	IssuerUrl     string
	RedirectUrl   string
	ListenAddress string
	ExtraScopes   []string
}

func (config *AuthProviderConfig) AddTokens(tokens *Tokens) {
	config.IdToken = tokens.IdToken
	config.RefreshToken = tokens.RefreshToken
}

// For logout
func (config *AuthProviderConfig) RemoveTokens() {
	config.IdToken = ""
	config.RefreshToken = ""
}

// Out
func (config AuthProviderConfig) ToKubeAuthProviderConfig() (authProviderConfig clientapi.AuthProviderConfig) {
	authProviderConfig.Config = make(map[string]string)
	authProviderConfig.Name = AuthProviderName
	authProviderConfig.Config[ParamIssuerUrl] = config.IssuerUrl
	authProviderConfig.Config[ParamRedirectUrl] = config.RedirectUrl
	authProviderConfig.Config[ParamClientId] = config.ClientId
	authProviderConfig.Config[ParamClientSecret] = config.ClientSecret
	authProviderConfig.Config[ParamIdToken] = config.IdToken
	authProviderConfig.Config[ParamRefreshToken] = config.RefreshToken

	if config.AuthMethod != DefaultAuthMethod && config.AuthMethod != "" {
		authProviderConfig.Config[ParamAuthMethod] = config.AuthMethod
	}
	if config.AuthRealm != "" {
		authProviderConfig.Config[ParamAuthRealm] = config.AuthRealm
	}
	if config.ListenAddress != DefaultListenAddress && config.ListenAddress != "" {
		authProviderConfig.Config[ParamListenAddress] = config.ListenAddress
	}
	if len(config.ExtraScopes) > 0 {
		authProviderConfig.Config[ParamExtraScopes] = strings.Join(config.ExtraScopes, ExtraScopesSeparator)
	}
	return
}

// In
func ProviderConfig(authConfig *clientapi.AuthProviderConfig) (config AuthProviderConfig, err error) {
	if clientId, exists := authConfig.Config[ParamClientId]; exists {
		config.ClientId = clientId
	} else {
		return config, fmt.Errorf("missing field in Auth Provider Config: %s", ParamClientId)
	}

	if clientSecret, exists := authConfig.Config[ParamClientSecret]; exists {
		config.ClientSecret = clientSecret
	} else {
		return config, fmt.Errorf("missing field in Auth Provider Config: %s", ParamClientSecret)
	}

	if issuerUrl, exists := authConfig.Config[ParamIssuerUrl]; exists {
		config.IssuerUrl = issuerUrl
	} else {
		return config, fmt.Errorf("missing field in Auth Provider Config: %s", issuerUrl)
	}

	if authMethod, exists := authConfig.Config[ParamAuthMethod]; exists {
		config.AuthMethod = authMethod
	}

	if authRealm, exists := authConfig.Config[ParamAuthRealm]; exists {
		config.AuthRealm = authRealm
	}

	if idToken, exists := authConfig.Config[ParamIdToken]; exists {
		config.IdToken = idToken
	}

	if refreshToken, exists := authConfig.Config[ParamRefreshToken]; exists {
		config.RefreshToken = refreshToken
	}

	if redirectUrl, exists := authConfig.Config[ParamRedirectUrl]; exists {
		config.RedirectUrl = redirectUrl
	}

	if extraScopes, exists := authConfig.Config[ParamExtraScopes]; exists {
		config.ExtraScopes = strings.Split(extraScopes, ExtraScopesSeparator)
	}

	return config, nil
}
