package oidc

// These are the minimal required fields to use kubectl's oidc auth plugin
type AuthProviderConfig struct {
	AuthMethod   string
	ClientId     string
	ClientSecret string
	IdToken      string
	RefreshToken string
	IssuerUrl    string
	RedirectUrl  string
	ExtraScopes  []string
}

type KubectlTokens struct {
	IdToken      string
	RefreshToken string
}

type BrowserAuthOptions struct {
	ListenAddress string
	ExtraParams   map[string]string
}
