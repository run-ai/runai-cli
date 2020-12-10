package oidc

// These are the minimal required fields to use kubectl's oidc auth plugin
type AuthProviderConfig struct {
	ClientId             string
	ClientSecret         string
	IdToken              string
	RefreshToken         string
	IssuerUrl            string
	RedirectUrl 		 string
	Scopes 				 []string
}

type KubectlTokens struct {
	IdToken      string
	RefreshToken string
}

type BrowserAuthOptions struct {
	ListenAddress string
	ExtraParams   map[string]string
}
