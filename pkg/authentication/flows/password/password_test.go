package password

import (
	"github.com/run-ai/runai-cli/pkg/authentication/kubeconfig"
	"gotest.tools/assert"
	"testing"
)

func TestConvertAuth0TokensToOauth2TokenAllFields(t *testing.T) {
	auth0Tokens := ServerTokens{
		RefreshToken: "refresh_test",
		IdToken:      "id_test",
	}

	oauth2Token := convertServerTokensToOauth2Token(&auth0Tokens)

	assert.Equal(t, oauth2Token.RefreshToken, "refresh_test")
	assert.Equal(t, oauth2Token.Extra(kubeconfig.IdTokenRawTokenName).(string), "id_test")
}

func TestConvertAuth0TokensToOauth2TokenOnlyRefreshToken(t *testing.T) {
	auth0Tokens := ServerTokens{
		RefreshToken: "refresh_test",
	}

	oauth2Token := convertServerTokensToOauth2Token(&auth0Tokens)

	assert.Equal(t, oauth2Token.RefreshToken, "refresh_test")
}

func TestConvertAuth0TokensToOauth2TokenOnlyIdToken(t *testing.T) {
	auth0Tokens := ServerTokens{
		IdToken: "id_test",
	}

	oauth2Token := convertServerTokensToOauth2Token(&auth0Tokens)

	assert.Equal(t, oauth2Token.Extra(kubeconfig.IdTokenRawTokenName).(string), "id_test")
}
