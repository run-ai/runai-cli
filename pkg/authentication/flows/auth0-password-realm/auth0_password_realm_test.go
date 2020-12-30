package auth0_password_realm

import (
	"github.com/run-ai/runai-cli/pkg/authentication/kubeconfig"
	"gotest.tools/assert"
	"testing"
)

func TestConvertAuth0TokensToOauth2TokenAllFields(t *testing.T) {
	auth0Tokens := auth0Tokens{
		RefreshToken: "refresh_test",
		IdToken:      "id_test",
	}

	oauth2Token := convertAuth0TokensToOauth2Token(&auth0Tokens)

	assert.Equal(t, oauth2Token.RefreshToken, "refresh_test")
	assert.Equal(t, oauth2Token.Extra(kubeconfig.IdTokenRawTokenName).(string), "id_test")
}

func TestConvertAuth0TokensToOauth2TokenOnlyRefreshToken(t *testing.T) {
	auth0Tokens := auth0Tokens{
		RefreshToken: "refresh_test",
	}

	oauth2Token := convertAuth0TokensToOauth2Token(&auth0Tokens)

	assert.Equal(t, oauth2Token.RefreshToken, "refresh_test")
}

func TestConvertAuth0TokensToOauth2TokenOnlyIdToken(t *testing.T) {
	auth0Tokens := auth0Tokens{
		IdToken: "id_test",
	}

	oauth2Token := convertAuth0TokensToOauth2Token(&auth0Tokens)

	assert.Equal(t, oauth2Token.Extra(kubeconfig.IdTokenRawTokenName).(string), "id_test")
}
