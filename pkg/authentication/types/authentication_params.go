package types

import "fmt"

const (
	CodePkceBrowser           = "browser"
	Auth0PasswordRealm        = "cli"
	defaultRedirectServer     = "localhost:8000"
	defaultAuthenticationFlow = CodePkceBrowser
)

type AuthenticationParams struct {
	ClientId      string
	IssuerURL     string
	ListenAddress string
	Auth0Realm    string

	AuthenticationFlow string
	User               string
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
	if a.Auth0Realm == "" {
		a.Auth0Realm = patch.Auth0Realm
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
	if a.ClientId == "" || a.IssuerURL == "" {
		return nil, fmt.Errorf("both client-id and idp-issuer-URL must be set")
	}
	if a.AuthenticationFlow == Auth0PasswordRealm && a.Auth0Realm == "" {
		return nil, fmt.Errorf("must provide auth0-realm when using CLI authentication")
	}
	return a, nil
}
