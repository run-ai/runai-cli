package password

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/run-ai/runai-cli/pkg/authentication/flows"
	"github.com/run-ai/runai-cli/pkg/authentication/kubeconfig"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/oauth2"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
)

const (
	mimeTypeUrlEncoded          = "application/x-www-form-urlencoded"
	auth0PasswordRealmGrantType = "http://auth0.com/oauth/grant-type/password-realm"
	keycloakPasswordGrantType   = "password"
)

type ServerTokens struct {
	RefreshToken string `json:"refresh_token,omitempty"`
	IdToken      string `json:"id_token,omitempty"`
}

func AuthenticateAuth0PasswordRealm(ctx context.Context, authParams *types.AuthenticationParams) (*oauth2.Token, error) {
	return sendAuthenticationRequest(ctx, auth0PasswordRealmGrantType, authParams.Auth0Realm, authParams)
}

func AuthenticateKeycloakPassword(ctx context.Context, authParams *types.AuthenticationParams) (*oauth2.Token, error) {
	return sendAuthenticationRequest(ctx, keycloakPasswordGrantType, "", authParams)
}

func sendAuthenticationRequest(ctx context.Context, grantType, realm string, authParams *types.AuthenticationParams) (*oauth2.Token, error) {
	user, password, err := getRawCredentials()
	if err != nil {
		return nil, err
	}

	requestParams := url.Values{
		"grant_type": {grantType},
		"username":   {user},
		"password":   {password},
		"scope":      {strings.Join(flows.Scopes, " ")},
		"client_id":  {authParams.ClientId},
	}

	if realm != "" {
		requestParams.Add("realm", realm)
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
	req.Header["Content-Type"] = []string{mimeTypeUrlEncoded}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		log.Debugf("invalid response: %v, %v, %v", res.StatusCode, res.Status, string(body))
		return nil, fmt.Errorf("invalid username or password")
	}
	contentType, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	return getTokenFromResponse(contentType, body)
}

func getTokenFromResponse(contentType string, responseBody []byte) (*oauth2.Token, error) {
	var token *oauth2.Token
	switch contentType {
	case mimeTypeUrlEncoded, "text/plain":
		formParams, err := url.ParseQuery(string(responseBody))
		if err != nil {
			return nil, err
		}
		token = (&oauth2.Token{
			TokenType:    formParams.Get("token_type"),
			RefreshToken: formParams.Get("refresh_token"),
		}).WithExtra(formParams)
		return token, nil
	default:
		var auth0Tokens ServerTokens
		if err := json.Unmarshal(responseBody, &auth0Tokens); err != nil {
			return nil, err
		}

		return convertServerTokensToOauth2Token(&auth0Tokens), nil
	}
}

func convertServerTokensToOauth2Token(auth0Tokens *ServerTokens) *oauth2.Token {
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

func readString(prompt string) (string, error) {
	if _, err := fmt.Fprint(os.Stderr, prompt); err != nil {
		return "", fmt.Errorf("write error: %v", err)
	}
	r := bufio.NewReader(os.Stdin)
	s, err := r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read error: %v", err)
	}
	s = strings.TrimRight(s, "\r\n")
	return s, nil
}

func readPassword(prompt string) (string, error) {
	if _, err := fmt.Fprint(os.Stderr, prompt); err != nil {
		return "", fmt.Errorf("write error: %v", err)
	}
	b, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", fmt.Errorf("read error: %v", err)
	}
	if _, err := fmt.Fprintln(os.Stderr); err != nil {
		return "", fmt.Errorf("write error: %v", err)
	}
	return string(b), nil
}
