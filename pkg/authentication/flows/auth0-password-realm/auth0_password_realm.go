package auth0_password_realm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/run-ai/runai-cli/pkg/authentication/flows"
	"github.com/run-ai/runai-cli/pkg/authentication/kubeconfig"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

const (
	Auth0PasswordRealmGrantType = "http://auth0.com/oauth/grant-type/password-realm"
	MimeTypeUrlEncoded          = "application/x-www-form-urlencoded"
)

type auth0Tokens struct {
	RefreshToken string `json:"refresh_token,omitempty"`
	IdToken      string `json:"id_token,omitempty"`
}

func AuthenticateAuth0PasswordRealm(ctx context.Context, authParams *types.AuthenticationParams) (*oauth2.Token, error) {
	user, password, err := getRawCredentials()
	if err != nil {
		return nil, err
	}

	requestParams := url.Values{
		"grant_type": {Auth0PasswordRealmGrantType},
		"realm":      {authParams.Auth0Realm},
		"username":   {user},
		"password":   {password},
		"scope":      {strings.Join(flows.Scopes, " ")},
		"client_id":  {authParams.ClientId},
	}
	provider, err := oidc.NewProvider(ctx, authParams.IssuerURL)
	if err != nil {
		return nil, err
	}

	log.Debug("Sending request to authentication")
	req, err := http.NewRequest("POST", provider.Endpoint().TokenURL, strings.NewReader(requestParams.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header["Content-Type"] = []string{MimeTypeUrlEncoded}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		log.Debugf("invalid response: %v, %v, %v", res.StatusCode, res.Status, string(body))
		return nil, fmt.Errorf("invalid username or password")
	}
	contentType, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	return getTokenFromROPCAuthResponse(contentType, body)
}

func getTokenFromROPCAuthResponse(contentType string, responseBody []byte) (*oauth2.Token, error) {
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
		return token, nil
	default:
		var auth0Tokens auth0Tokens
		if err := json.Unmarshal(responseBody, &auth0Tokens); err != nil {
			return nil, err
		}

		return convertAuth0TokensToOauth2Token(&auth0Tokens), nil
	}
}

func convertAuth0TokensToOauth2Token(auth0Tokens *auth0Tokens) *oauth2.Token {
	oauth2Token := &oauth2.Token{
		RefreshToken: auth0Tokens.RefreshToken,
	}
	extraTokensOauth2 := make(map[string]interface{})
	extraTokensOauth2[kubeconfig.IdTokenRawTokenName] = auth0Tokens.IdToken
	oauth2Token = oauth2Token.WithExtra(extraTokensOauth2)
	return oauth2Token
}

func getRawCredentials() (string, string, error) {
	var err error
	var username, password string
	if username, err = readString("Username: "); err != nil {
		return "", "", err
	}
	if password, err = readPassword("Password: "); err != nil {
		return "", "", err
	}
	return username, password, nil
}
