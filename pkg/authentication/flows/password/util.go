package password

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/run-ai/runai-cli/pkg/authentication/kubeconfig"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/oauth2"
	"net/url"
	"os"
	"strings"
	"syscall"
)

const (
	MimeTypeUrlEncoded = "application/x-www-form-urlencoded"
)

type ServerTokens struct {
	RefreshToken string `json:"refresh_token,omitempty"`
	IdToken      string `json:"id_token,omitempty"`
}

func GetTokenFromResponse(contentType string, responseBody []byte) (*oauth2.Token, error) {
	var token *oauth2.Token
	switch contentType {
	case MimeTypeUrlEncoded, "text/plain":
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

		return ConvertServerTokensToOauth2Token(&auth0Tokens), nil
	}
}

func ConvertServerTokensToOauth2Token(auth0Tokens *ServerTokens) *oauth2.Token {
	oauth2Token := &oauth2.Token{
		RefreshToken: auth0Tokens.RefreshToken,
	}
	extraTokensOauth2 := make(map[string]interface{})
	extraTokensOauth2[kubeconfig.IdTokenRawTokenName] = auth0Tokens.IdToken
	oauth2Token = oauth2Token.WithExtra(extraTokensOauth2)
	return oauth2Token
}

func GetRawCredentials() (string, string, error) {
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
