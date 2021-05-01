package flows

import (
	"context"
	"github.com/coreos/go-oidc"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	"golang.org/x/oauth2"
)

const (
	OpenIdScope       = "openid"
	RefreshTokenScope = "offline_access"
	EmailScope        = "email"
)

var Scopes = []string{EmailScope, OpenIdScope, RefreshTokenScope}

func GetOauth2Config(ctx context.Context, authParams *types.AuthenticationParams) (*oauth2.Config, error) {
	provider, err := oidc.NewProvider(ctx, authParams.IssuerURL)
	if err != nil {
		return nil, err
	}
	return &oauth2.Config{
		ClientID:    authParams.ClientId,
		Endpoint:    provider.Endpoint(),
		Scopes:      Scopes,
		RedirectURL: authParams.ListenAddress,
	}, nil
}
