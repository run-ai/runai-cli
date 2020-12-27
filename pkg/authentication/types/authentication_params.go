package types

import "fmt"

const (
	CodePkceBrowser           = "browser"
	defaultRedirectServer     = "localhost:3000"
	defaultAuthenticationFlow = CodePkceBrowser
)

type AuthenticationParams struct {
	ClientId      string
	IssuerURL     string
	ListenAddress string

	AuthenticationFlow string
}

func (a *AuthenticationParams) GetRedirectUrl() string {
	return fmt.Sprintf("http://%s/", a.ListenAddress)
}

func (a *AuthenticationParams) MergeAuthenticationParams(patch *AuthenticationParams) *AuthenticationParams {
	if a.ClientId == "" {
		a.ClientId = patch.ClientId
	}
	if a.IssuerURL == "" {
		a.IssuerURL = patch.IssuerURL
	}
	if a.ListenAddress == "" {
		a.ListenAddress = patch.ListenAddress
	}
	if a.AuthenticationFlow == "" {
		a.AuthenticationFlow = patch.AuthenticationFlow
	}
	return a
}

func (a *AuthenticationParams) ValidateAndSetDefaultAuthenticationParams() (*AuthenticationParams, error) {
	if a.ListenAddress == "" {
		a.ListenAddress = defaultRedirectServer
	}
	if a.AuthenticationFlow == "" {
		a.AuthenticationFlow = defaultAuthenticationFlow
	}
	if a.ClientId == "" && a.IssuerURL == "" {
		return nil, fmt.Errorf("both Client-id and Issuer-URL must be set")
	}
	return a, nil
}
