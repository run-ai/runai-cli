package code_pkce_browser

import (
	"context"
	"github.com/int128/oauth2cli"
	"github.com/pkg/browser"
	"github.com/run-ai/runai-cli/pkg/authentication/flows"
	"github.com/run-ai/runai-cli/pkg/authentication/pages"
	"github.com/run-ai/runai-cli/pkg/authentication/pkce"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func AuthenticateCodePkceBrowser(ctx context.Context, authParams *types.AuthenticationParams) (*oauth2.Token, error) {
	log.Debug("Authentication process start with authorization code flow, with PKCE, browser mode")
	localServerReadyChan := make(chan string, 1)
	go waitForLocalServer(localServerReadyChan)

	oauth2Config, err := flows.GetOauth2Config(ctx, authParams)
	if err != nil {
		return nil, err
	}
	oauth2Config.Scopes = append(oauth2Config.Scopes, authParams.AdditionalScopes...)
	log.Debugf("Generated oauth2config object: %v", oauth2Config)
	oauth2cliConfig, err := getOauth2cliGetTokenConfig(oauth2Config, localServerReadyChan, authParams)
	if err != nil {
		return nil, err
	}
	log.Debug("Generated oauth2cli object")
	return oauth2cli.GetToken(ctx, *oauth2cliConfig)
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
		LocalServerSuccessHTML: pages.LoginPageHtml,
	}, nil
}

func waitForLocalServer(readyChan chan string) {
	url := <-readyChan
	log.Debugf("Opening browser to URL: %v", url)
	browser.OpenURL(url)
}
