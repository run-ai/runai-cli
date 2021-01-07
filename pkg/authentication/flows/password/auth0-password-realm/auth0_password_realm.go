package auth0_password_realm

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/run-ai/runai-cli/pkg/authentication/flows"
	passwordFlow "github.com/run-ai/runai-cli/pkg/authentication/flows/password"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

const (
	auth0PasswordRealmGrantType = "http://auth0.com/oauth/grant-type/password-realm"
)

func AuthenticateAuth0PasswordRealm(ctx context.Context, authParams *types.AuthenticationParams) (*oauth2.Token, error) {
	user, password, err := passwordFlow.GetRawCredentials()
	if err != nil {
		return nil, err
	}

	requestParams := url.Values{
		"grant_type": {auth0PasswordRealmGrantType},
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
	req.Header["Content-Type"] = []string{passwordFlow.MimeTypeUrlEncoded}
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
		return nil, fmt.Errorf("Invalid username or password")
	}
	contentType, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	return passwordFlow.GetTokenFromResponse(contentType, body)
}
