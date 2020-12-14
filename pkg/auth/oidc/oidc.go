package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/int128/oauth2cli"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	. "github.com/run-ai/runai-cli/pkg/auth/config"
	"github.com/run-ai/runai-cli/pkg/auth/util"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"

	gooidc "github.com/coreos/go-oidc"
)

const (
	Auth0PasswordRealmGrantType = "http://auth0.com/oauth/grant-type/password-realm"
	MimeTypeUrlEncoded          = "application/x-www-form-urlencoded"
)

var DefaultScopes = []string{"email", gooidc.ScopeOpenID, gooidc.ScopeOfflineAccess}

type Authenticator struct {
	provider  *gooidc.Provider
	config    oauth2.Config
	issuerUrl string
	ctx       context.Context

	// Although its possible to make many requests using a single Authenticator, in practice it is only used once per command thus we put these 2 params here as a convenience
	// BUT if you ever find yourself using the same Authenticator instance more then once you must pass a new nonce and state each time.
	state     string
	nonce     string
	authRealm string
}

func NewAuthenticator(config AuthProviderConfig) (*Authenticator, error) {
	ctx := context.Background()
	provider, err := gooidc.NewProvider(ctx, config.IssuerUrl)
	if err != nil {
		return nil, err
	}
	conf := oauth2.Config{
		ClientID:     config.ClientId,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectUrl,
		Endpoint:     provider.Endpoint(),
		Scopes:       util.MergeScopes(DefaultScopes, config.ExtraScopes),
	}
	return &Authenticator{
		provider:  provider,
		config:    conf,
		issuerUrl: config.IssuerUrl,
		ctx:       ctx,
		state:     util.MakeNonce(),
		nonce:     util.MakeNonce(),
		authRealm: config.AuthRealm,
	}, nil
}

// When a browser is locally available, opens a browser automatically,listens for the response and gets a token with the code in the response.
func (authenticator Authenticator) BrowserAuth(options BrowserAuthOptions) (*Tokens, error) {
	var (
		readyChan = make(chan string, 1)
		eg        errgroup.Group
		tokens    *Tokens
	)
	eg.Go(func() error {
		select {
		case authUrl, ok := <-readyChan:
			if !ok {
				return nil
			}
			log.Debugf("opening %s in the browser", authUrl)
			if err := browser.OpenURL(authUrl); err != nil {
				err = errors.Wrap(err, "Could not open browser")
				log.Error(err)
				return err
			}
			return nil
		case <-authenticator.ctx.Done():
			return fmt.Errorf("context cancelled while waiting for the local server: %v", authenticator.ctx.Err())
		}
	})

	eg.Go(func() error {
		defer close(readyChan)
		token, err := oauth2cli.GetToken(authenticator.ctx, authenticator.buildCliConfig(options, readyChan))
		if err != nil {
			return fmt.Errorf("error during auth code flow: %v", err)
		}
		verified, err := authenticator.verifyAndConvertToken(token)
		tokens = verified
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("error encountered during authentication: %v", err)
	}
	return tokens, nil
}

// When a browser is not locally available (i.e. an ssh session) shows the url for the user to put in their remotely-available browser and prompts for the code.
func (authenticator Authenticator) RemoteBrowserAuth() (*Tokens, error) {
	// TODO [by dan]: pass extras
	authUrl := authenticator.config.AuthCodeURL(authenticator.state, authenticator.authRequestOptions(make(map[string]string))...)
	fmt.Printf("Please go to this url in any browser: %s \n\n", authUrl)
	code, err := util.ReadPassword("And paste the code here: ")
	if err != nil {
		return &Tokens{}, err
	}

	// TODO [by dan]: pass extras
	if token, err := authenticator.config.Exchange(authenticator.ctx, code, authenticator.authRequestOptions(make(map[string]string))...); err == nil {
		return authenticator.verifyAndConvertToken(token)
	} else {
		return &Tokens{}, err
	}
}

// To be able to use separate connections on separate applications, auth0 requires passing a non-standard grant type and a non-standard scope to tell it what connection to use for
// the authentication request.
// There is absolutely 0 support for such a custom case in the oauth2 standard library so we're forced to make this call manually.
func (authenticator Authenticator) Auth0PasswordAuth() (*Tokens, error) {
	// Auth0 has an unfortunate way of making the user/connection separation works so if a realm has been passed we need to use their api accordingly.
	// See more here: https://auth0.com/docs/flows/call-your-api-using-resource-owner-password-flow#configure-realm-support
	if authenticator.authRealm == "" {
		// If no realm is passed then this cam be handled as a standard ROPC request, auth0 requires that you set a default connection for the entire tenant and it will be used
		// to authenticate all requests. For instance our dev and staging tenants are structured like this.
		return authenticator.PasswordAuth()
	}
	var err error
	var username, password string
	if username, err = util.ReadString("Username: "); err != nil {
		return nil, err
	}
	if password, err = util.ReadPassword("Password: "); err != nil {
		return nil, err
	}
	return authenticator.auth0ROPC(username, password)
}

func (authenticator Authenticator) auth0ROPC(username string, password string) (*Tokens, error) {
	var req *http.Request
	var res *http.Response
	var err error
	requestParams := url.Values{
		"grant_type":    {Auth0PasswordRealmGrantType},
		"realm":         {authenticator.authRealm},
		"username":      {username},
		"password":      {password},
		"scope":         {strings.Join(authenticator.config.Scopes, " ")},
		"client_id":     {authenticator.config.ClientID},
		"client_secret": {authenticator.config.ClientSecret},
	}
	if req, err = http.NewRequest("POST", authenticator.provider.Endpoint().TokenURL, strings.NewReader(requestParams.Encode())); err != nil {
		return nil, err
	}
	req.Header["Content-Type"] = []string{MimeTypeUrlEncoded}
	if res, err = http.DefaultClient.Do(req); err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("bad auth response --> %d : %s / %s", res.StatusCode, res.Status, string(body))
	}
	contentType, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	return authenticator.getTokenFromROPCAuthResponse(contentType, body)
}

func (authenticator Authenticator) getTokenFromROPCAuthResponse(contentType string, responseBody []byte) (*Tokens, error) {
	var token *oauth2.Token
	switch contentType {
	case MimeTypeUrlEncoded, "text/plain":
		formParams, err := url.ParseQuery(string(responseBody))
		if err != nil {
			return nil, err
		}
		token = (&oauth2.Token{
			TokenType:    formParams.Get("token_type"),
			AccessToken:  formParams.Get("access_token"),
			RefreshToken: formParams.Get("refresh_token"),
		}).WithExtra(formParams)
		return authenticator.verifyAndConvertToken(token)
	default:
		tokens := &Tokens{}
		if err := json.Unmarshal(responseBody, tokens); err != nil {
			return nil, err
		}
		return tokens, authenticator.verifyToken(tokens.IdToken)
	}
}

// Resource Owner Password Credentials flow, supported by Keycloak (and probably any other IDP)
func (authenticator Authenticator) PasswordAuth() (tokens *Tokens, err error) {
	var username, password string
	if username, err = util.ReadString("Username: "); err != nil {
		return nil, err
	}
	if password, err = util.ReadPassword("Password: "); err != nil {
		return nil, err
	}
	if token, err := authenticator.config.PasswordCredentialsToken(authenticator.ctx, username, password); err == nil {
		return authenticator.verifyAndConvertToken(token)
	} else {
		return nil, err
	}
}

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

func (authenticator Authenticator) buildCliConfig(options BrowserAuthOptions, readyChan chan string) oauth2cli.Config {
	cliConfig := oauth2cli.Config{
		OAuth2Config:           authenticator.config,
		State:                  authenticator.state,
		AuthCodeOptions:        authenticator.authRequestOptions(options.ExtraParams),
		LocalServerBindAddress: []string{options.ListenAddress},
		LocalServerReadyChan:   readyChan,
		RedirectURLHostname:    "localhost", // Can be made configurable, if needed
		Logf:                   log.Debugf,
		//TODO LocalServerSuccessHTML: ,  --> Can be configured however we want (messages, links, logo etc.)
	}
	return cliConfig
}

// verifyAndConvertToken calls verifyToken and converts the oauth2 token to our representation of token
func (authenticator Authenticator) verifyAndConvertToken(token *oauth2.Token) (*Tokens, error) {
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("id_token is missing in the token response: %v", token)
	}
	if err := authenticator.verifyToken(idToken); err != nil {
		return nil, err
	}
	return &Tokens{
		IdToken:      idToken,
		RefreshToken: token.RefreshToken,
	}, nil
}

// verifyToken verifies the token with the certificates of the provider and the nonce.
// If the nonce is an empty string, it does not verify the nonce.
func (authenticator Authenticator) verifyToken(idToken string) error {
	verifier := authenticator.provider.Verifier(&gooidc.Config{ClientID: authenticator.config.ClientID, Now: time.Now})
	verifiedIDToken, err := verifier.Verify(authenticator.ctx, idToken)
	if err != nil {
		return fmt.Errorf("error while verifying ID token: %v", err)
	}
	if verifiedIDToken.Nonce != "" && (authenticator.nonce != verifiedIDToken.Nonce) {
		return fmt.Errorf("token nonce did not match! (expected %s but got %s)", authenticator.nonce, verifiedIDToken.Nonce)
	}
	return nil
}
