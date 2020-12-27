package authentication

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/int128/oauth2cli"
	"github.com/run-ai/runai-cli/pkg/authentication/pkce"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	"golang.org/x/oauth2"
)

func authenticateCodePkceBrowser(ctx context.Context, authParams *types.AuthenticationParams) (*oauth2.Token, error) {
	localServerReadyChan := make(chan string, 1)
	go waitForLocalServer(localServerReadyChan)

	oauth2Config, err := getOauth2Config(ctx, authParams)
	if err != nil {
		return nil, err
	}
	oauth2cliConfig, err := getOauth2cliGetTokenConfig(oauth2Config, localServerReadyChan, authParams)
	if err != nil {
		return nil, err
	}
	token, err := oauth2cli.GetToken(ctx, *oauth2cliConfig)
	return token, nil
}

func getOauth2cliGetTokenConfig(oauth2Config *oauth2.Config, localServerReadyChan chan string, authParams *types.AuthenticationParams) (*oauth2cli.Config, error) {
	pkceParams, err := pkce.New()
	if err != nil {
		return nil, err
	}
	authCodeOptions := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam(pkce.CodeChallengeParamName, pkceParams.CodeChallenge),
		oauth2.SetAuthURLParam(pkce.CodeChallengeMethodParamName, pkceParams.CodeChallengeMethod),
	}
	tokenRequestOptions := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam(pkce.CodeVerifierParamName, pkceParams.CodeVerifier),
	}

	return &oauth2cli.Config{
		OAuth2Config:           *oauth2Config,
		LocalServerBindAddress: []string{authParams.ListenAddress},
		LocalServerReadyChan:   localServerReadyChan,
		AuthCodeOptions:        authCodeOptions,
		TokenRequestOptions:    tokenRequestOptions,
	}, nil
}

func getOauth2Config(ctx context.Context, authParams *types.AuthenticationParams) (*oauth2.Config, error) {
	provider, err := oidc.NewProvider(ctx, authParams.IssuerURL)
	if err != nil {
		return nil, err
	}
	return &oauth2.Config{
		ClientID:    authParams.ClientId,
		Endpoint:    provider.Endpoint(),
		Scopes:      []string{openIdScope, refreshTokenScope},
		RedirectURL: authParams.ListenAddress,
	}, nil
}

func waitForLocalServer(readyChan chan string) {
	url := <-readyChan
	fmt.Printf("You can go to %v \n", url)
}
