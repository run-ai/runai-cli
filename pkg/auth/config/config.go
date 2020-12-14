package config

import (
	"fmt"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	"strings"
)

const (
	ClientId             = "client-id"
	ClientSecret         = "client-secret"
	IdToken              = "id-token"
	RefreshToken         = "refresh-token"
	IssuerUrl            = "idp-issuer-url"
	RedirectUrl          = "redirect-url"
	AuthProviderName     = "oidc"
	AuthMethod           = "auth-method"
	ExtraScopes          = "auth-request-extra-scopes"
	AuthRealm            = "auth-realm"
	ListenAddress        = "listen-address"
	ExtraScopesSeparator = ","

	DefaultKubeConfigUserName = "runai-oidc"
	DefaultListenAddress      = "127.0.0.1:8000"
	DefaultIssuerUrl          = "https://runai-prod.auth0.com/"
	DefaultRedirectUrl        = "https://app.run.ai/auth"
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
	ExtraParams   map[string]string
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
	authProviderConfig.Config[ClientId] = config.ClientId
	authProviderConfig.Config[ClientSecret] = config.ClientSecret
	authProviderConfig.Config[IdToken] = config.IdToken
	authProviderConfig.Config[RefreshToken] = config.RefreshToken
	if config.AuthMethod != DefaultAuthMethod && config.AuthMethod != "" {
		authProviderConfig.Config[AuthMethod] = config.AuthMethod
	}
	if config.AuthRealm != "" {
		authProviderConfig.Config[AuthRealm] = config.AuthRealm
	}
	if config.IssuerUrl != DefaultIssuerUrl && config.IssuerUrl != "" {
		authProviderConfig.Config[IssuerUrl] = config.IssuerUrl
	}
	if config.RedirectUrl != DefaultRedirectUrl && config.RedirectUrl != "" {
		authProviderConfig.Config[RedirectUrl] = config.RedirectUrl
	}
	if config.ListenAddress != DefaultListenAddress && config.ListenAddress != "" {
		authProviderConfig.Config[ListenAddress] = config.ListenAddress
	}
	if len(config.ExtraScopes) > 0 {
		authProviderConfig.Config[ExtraScopes] = strings.Join(config.ExtraScopes, ExtraScopesSeparator)
	}
	return
}

// In
func ProviderConfig(authConfig *clientapi.AuthProviderConfig) (config AuthProviderConfig, err error) {
	if clientId, exists := authConfig.Config[ClientId]; exists {
		config.ClientId = clientId
	} else {
		return config, fmt.Errorf("missing field in Auth Provider Config: %s", ClientId)
	}

	if clientSecret, exists := authConfig.Config[ClientSecret]; exists {
		config.ClientSecret = clientSecret
	} else {
		return config, fmt.Errorf("missing field in Auth Provider Config: %s", ClientSecret)
	}

	if issuerUrl, exists := authConfig.Config[IssuerUrl]; exists {
		config.IssuerUrl = issuerUrl
	} else {
		return config, fmt.Errorf("missing field in Auth Provider Config: %s", issuerUrl)
	}

	if authMethod, exists := authConfig.Config[AuthMethod]; exists {
		config.AuthMethod = authMethod
	}

	if authRealm, exists := authConfig.Config[AuthRealm]; exists {
		config.AuthMethod = authRealm
	}

	if idToken, exists := authConfig.Config[IdToken]; exists {
		config.IdToken = idToken
	}

	if refreshToken, exists := authConfig.Config[RefreshToken]; exists {
		config.RefreshToken = refreshToken
	}

	if redirectUrl, exists := authConfig.Config[RedirectUrl]; exists {
		config.RedirectUrl = redirectUrl
	}

	if extraScopes, exists := authConfig.Config[ExtraScopes]; exists {
		config.ExtraScopes = strings.Split(extraScopes, ExtraScopesSeparator)
	}

	return config, nil
}
