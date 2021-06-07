package types

import (
	"fmt"
	"github.com/run-ai/runai-cli/cmd/util"
)

const (
	CodePkceBrowser           = "browser"
	CodePkceRemoteBrowser     = "remote-browser"
	ClientCredentials         = "cli"
	defaultRedirectServer     = "localhost:8000"
	defaultAirgappedFlag      = false
	defaultAuthenticationFlow = CodePkceBrowser
)

type AuthenticationParams struct {
	ClientId         string
	IssuerURL        string
	ListenAddress    string
	Realm            string
	AdditionalScopes []string

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
	if a.Realm == "" {
		a.Realm = patch.Realm
	}
	if a.IsAirgapped == nil {
		a.IsAirgapped = patch.IsAirgapped
	}
	if len(patch.AdditionalScopes) != 0 {
		a.AdditionalScopes = append(a.AdditionalScopes, patch.AdditionalScopes...)
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
	if a.AuthenticationFlow == ClientCredentials && a.Realm == "" && !util.IsBoolPTrue(a.IsAirgapped) {
		return nil, fmt.Errorf("must provide realm when using CLI authentication")
	}
	return a, nil
}
