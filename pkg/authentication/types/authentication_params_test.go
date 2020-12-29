package types

import (
	"gotest.tools/assert"
	"testing"
)

func TestMergeAuthenticationParams_CliWins(t *testing.T) {
	cliAuthenticationParams := &AuthenticationParams{
		ClientId:      "testClientId",
		IssuerURL:     "testIssuerUrl",
		ListenAddress: "testListenAddress",
		User:          "testUser",
	}

	kubeConfigAuthenticationParams := &AuthenticationParams{
		ClientId:      "badClientId",
		IssuerURL:     "badIssuerUrl",
		ListenAddress: "badListenAddress",
	}

	result := cliAuthenticationParams.MergeAuthenticationParams(kubeConfigAuthenticationParams)

	assert.Equal(t, result.ClientId, "testClientId")
	assert.Equal(t, result.IssuerURL, "testIssuerUrl")
	assert.Equal(t, result.ListenAddress, "testListenAddress")
	assert.Equal(t, result.User, "testUser")
}

func TestMergeAuthenticationParams_KubeConfigEmpty(t *testing.T) {
	cliAuthenticationParams := &AuthenticationParams{
		ClientId:      "testClientId",
		IssuerURL:     "testIssuerUrl",
		ListenAddress: "testListenAddress",
		User:          "testUser",
	}

	kubeConfigAuthenticationParams := &AuthenticationParams{}

	result := cliAuthenticationParams.MergeAuthenticationParams(kubeConfigAuthenticationParams)

	assert.Equal(t, result.ClientId, "testClientId")
	assert.Equal(t, result.IssuerURL, "testIssuerUrl")
	assert.Equal(t, result.ListenAddress, "testListenAddress")
	assert.Equal(t, result.User, "testUser")
}

func TestMergeAuthenticationParams_CliEmpty(t *testing.T) {
	kubeConfigAuthenticationParams := &AuthenticationParams{
		ClientId:      "testClientId",
		IssuerURL:     "testIssuerUrl",
		ListenAddress: "testListenAddress",
	}

	cliAuthenticationParams := &AuthenticationParams{}

	result := cliAuthenticationParams.MergeAuthenticationParams(kubeConfigAuthenticationParams)

	assert.Equal(t, result.ClientId, "testClientId")
	assert.Equal(t, result.IssuerURL, "testIssuerUrl")
	assert.Equal(t, result.ListenAddress, "testListenAddress")
}

func TestValidateAndSetDefaultAuthenticationParams_allSetOK(t *testing.T) {
	authenticationParams := &AuthenticationParams{
		ClientId:           "testClientId",
		IssuerURL:          "testIssuerUrl",
		ListenAddress:      "testListenAddress",
		User:               "testUser",
		AuthenticationFlow: defaultAuthenticationFlow,
	}

	result, err := authenticationParams.ValidateAndSetDefaultAuthenticationParams()

	assert.Equal(t, err, nil)
	assert.Equal(t, result.ClientId, "testClientId")
	assert.Equal(t, result.IssuerURL, "testIssuerUrl")
	assert.Equal(t, result.ListenAddress, "testListenAddress")
	assert.Equal(t, result.User, "testUser")
	assert.Equal(t, result.AuthenticationFlow, defaultAuthenticationFlow)
}

func TestValidateAndSetDefaultAuthenticationParams_noClientId(t *testing.T) {
	authenticationParams := &AuthenticationParams{
		IssuerURL:          "testIssuerUrl",
		ListenAddress:      "testListenAddress",
		User:               "testUser",
		AuthenticationFlow: defaultAuthenticationFlow,
	}

	_, err := authenticationParams.ValidateAndSetDefaultAuthenticationParams()

	if err == nil {
		t.FailNow()
	}
}
