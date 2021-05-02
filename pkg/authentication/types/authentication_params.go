package types

import (
	"fmt"
	"github.com/run-ai/runai-cli/cmd/util"
)

const (
	CodePkceBrowser           = "browser"
	CodePkceRemoteBrowser	  = "remote-browser"
	Auth0PasswordRealm        = "cli"
	defaultRedirectServer     = "localhost:8000"
	defaultAirgappedFlag      = false
	defaultAuthenticationFlow = CodePkceBrowser
)

type AuthenticationParams struct {
	ClientId      string
	IssuerURL     string
	ListenAddress string
	Auth0Realm    string

	AuthenticationFlow string
	User               string
	IsAirgapped        *bool
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
	if a.IsAirgapped == nil {
		a.IsAirgapped = patch.IsAirgapped
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
	if a.IsAirgapped == nil {
		defaultValue := defaultAirgappedFlag
		a.IsAirgapped = &defaultValue
	}
	if a.ClientId == "" || a.IssuerURL == "" {
		return nil, fmt.Errorf("both client-id and idp-issuer-URL must be set")
	}
	if a.AuthenticationFlow == Auth0PasswordRealm && a.Auth0Realm == "" && !util.IsBoolPTrue(a.IsAirgapped) {
		return nil, fmt.Errorf("must provide auth0-realm when using CLI authentication")
	}
	return a, nil
}
