package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/int128/oauth2cli"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"time"

	"golang.org/x/oauth2"

	gooidc "github.com/coreos/go-oidc"
)

type Authenticator struct {
	provider *gooidc.Provider
	config   oauth2.Config
	ctx      context.Context

	// Although its possible to make many requests using a single Authenticator, in practice it is only used once per command thus we put these 2 params here as a convenience
	// BUT if you ever find yourself using the same Authenticator instance more then once you must pass a new nonce and state each time.
	state string
	nonce string
}

func NewAuthenticator(config AuthProviderConfig) (*Authenticator, error) {
	ctx := context.Background()

	provider, err := gooidc.NewProvider(ctx, config.IssuerUrl)
	if err != nil {
		log.Infof("failed to create provider: %v", err)
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     config.ClientId,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectUrl,
		Endpoint:     provider.Endpoint(),
		Scopes:       config.Scopes,
	}

	return &Authenticator{
		provider: provider,
		config:   conf,
		ctx:      ctx,
		state:    makeNonce(),
		nonce:    makeNonce(),
	}, nil
}

func (authenticator Authenticator) ToAuthProviderConfig(tokens *KubectlTokens) (config clientcmdapi.AuthProviderConfig) {
	config.Config[ClientId] = authenticator.config.ClientID
	config.Config[ClientSecret] = authenticator.config.ClientSecret
	config.Config[IssuerUrl] = authenticator.config.Endpoint.AuthURL
	config.Config[IdToken] = tokens.IdToken
	config.Config[RefreshToken] = tokens.RefreshToken
	return
}


// When a browser is locally available, opens a browser automatically,listens for the response and gets a token with the code in the response.
func (authenticator Authenticator) BrowserAuth(options BrowserAuthOptions) (*KubectlTokens, error) {

	// TODO [by dan]: this is the logic kubelogin does to 'lock' the listen port - not sure if its actually required?

	//_, port, err := net.SplitHostPort(options.ListenAddress)
	//if err != nil {
	//	log.Errorf("Bad listen address given: %w", err)
	//	return nil, err
	//}
	//log.Debugf("Trying to lock local port %s", port)

	var (
		readyChan = make(chan string, 1)
		eg        errgroup.Group
		tokens    *KubectlTokens
	)
	eg.Go(func() error {
		select {
		case url, ok := <-readyChan:
			if !ok {
				return nil
			}
			log.Infof("opening %s in the browser", url)
			if err := browser.OpenURL(url); err != nil {
				err = errors.Wrap(err, "Could not open browser")
				log.Error(err)
				return err
			}
			return nil
		case <-authenticator.ctx.Done():
			return fmt.Errorf("context cancelled while waiting for the local server: %w", authenticator.ctx.Err())
		}
	})

	cliConfig := authenticator.buildCliConfig(options, readyChan)

	eg.Go(func() error {
		defer close(readyChan)
		token, err := oauth2cli.GetToken(authenticator.ctx, cliConfig)
		if err != nil {
			return fmt.Errorf("error during auth code flow: %w", err)
		}
		verified, err := authenticator.verifyToken(authenticator.ctx, token)
		tokens = verified
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("error encountered during authentication: %w", err)
	}
	return tokens, nil
}

// When a browser is not locally available (i.e. an ssh session) shows the url for the user to put in their remotely-available browser and prompts for the code.
/*func (authenticator Authenticator) RemoteBrowserAuth() KubectlTokens {

}*/

// To be able to use separate connections on separate applications, auth0 requires passing a non-standard grant type and a non-standard scope to tell it what connection to use for
// the authentication request.
/*func (authenticator Authenticator) Auth0PasswordAuth() KubectlTokens {

}*/

// standard Resource Owner Password Credentials flow, supported by Keycloak
/*func (authenticator Authenticator) PasswordAuth() KubectlTokens {

}*/

/*func (authenticator Authenticator) authCodeUrl(extraParams map[string]string) string {
	requestOpts := authenticator.authRequestOptions(extraParams)
	return authenticator.config.AuthCodeURL(authenticator.state, requestOpts...)
}

func (authenticator Authenticator) exchangeCodeForToken(code string) KubectlTokens {
	// TODO [by dan]: verify state!

	oauth2Token, err := authenticator.config.Exchange(authenticator.ctx, code)
	if err != nil {
		// handle error
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		// handle missing token
	}

	verifier := authenticator.provider.Verifier(&gooidc.Config{ClientID: authenticator.config.ClientID})
	// Parse and verify ID Token payload.
	idToken, err := verifier.Verify(authenticator.ctx, rawIDToken)
	if err != nil {
		// handle error
	}

	// TODO [by dan]: since we can see the email claim here we can potentially write the user name into kubeconfig and use that instead of the token parsing logic I put here
	// TODO [by dan]: second option: drop a .userName file in .runai and hold username there
	// TODO [by dan]: also kubelogin's authentication package does time verification
	// Extract custom claims
	var claims struct {
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		// handle error
	}
}*/

func (authenticator Authenticator) authRequestOptions(extraParams map[string]string) []oauth2.AuthCodeOption {
	options := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline, // Required to get a refresh token with the response
		gooidc.Nonce(authenticator.nonce),
	}
	for key, value := range extraParams {
		options = append(options, oauth2.SetAuthURLParam(key, value))
	}
	return options
}

func makeNonce() string {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		log.Debug(err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buffer)

}

func (authenticator Authenticator) buildCliConfig(options BrowserAuthOptions, readyChan chan string) oauth2cli.Config {
	cliConfig := oauth2cli.Config{
		OAuth2Config:           authenticator.config,
		State:                  authenticator.state,
		AuthCodeOptions:        authenticator.authRequestOptions(options.ExtraParams),
		LocalServerBindAddress: []string{options.ListenAddress},
		LocalServerReadyChan:   readyChan,
		RedirectURLHostname:    "localhost",          // Can be made configurable, if needed
		LocalServerSuccessHTML: RedirectHTMLResponse, // Can be configured however we want (messages, links, logo etc.)
		Logf:                   log.Debugf,
	}
	return cliConfig
}

// verifyToken verifies the token with the certificates of the provider and the nonce.
// If the nonce is an empty string, it does not verify the nonce.
func (authenticator Authenticator) verifyToken(ctx context.Context, token *oauth2.Token) (*KubectlTokens, error) {
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("id_token is missing in the token response: %s", token)
	}
	verifier := authenticator.provider.Verifier(&gooidc.Config{ClientID: authenticator.config.ClientID, Now: time.Now})
	verifiedIDToken, err := verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("error while verifying ID token: %w", err)
	}
	if authenticator.nonce != verifiedIDToken.Nonce {
		return nil, fmt.Errorf("makeNonce did not match (expected %s but got %s)", authenticator.nonce, verifiedIDToken.Nonce)
	}
	return &KubectlTokens{
		IdToken:      idToken,
		RefreshToken: token.RefreshToken,
		// Theres also token.AccessToken, but it seems the kubectl oidc authenticator knows how to refresh without it.
	}, nil
}
