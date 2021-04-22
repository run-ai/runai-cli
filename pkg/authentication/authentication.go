package authentication

import (
	"context"
	"fmt"
	"github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/authentication/flows/code-pkce-browser"
	"github.com/run-ai/runai-cli/pkg/authentication/flows/password"
	"github.com/run-ai/runai-cli/pkg/authentication/jwt"
	"github.com/run-ai/runai-cli/pkg/authentication/kubeconfig"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func GetCurrentAuthenticateUser() (string, error) {
	idToken, err := kubeconfig.GetCurrentUserIdToken()
	if err != nil {
		return "", err
	}

	token, err := jwt.Decode(idToken)
	if err != nil {
		return "", err
	}
	return token.Email, nil
}

func GetCurrentAuthenticateUserSubject() (string, string, error) {
	idToken, err := kubeconfig.GetCurrentUserIdToken()
	if err != nil {
		return "", "", err
	}

	token, err := jwt.Decode(idToken)
	if err != nil {
		return "", "", err
	}
	return token.Subject, token.Email, nil
}

func GetCurrentAuthenticateUserUidGid() (string, string, error) {
	idToken, err := kubeconfig.GetCurrentUserIdToken()
	if err != nil {
		return "", "", err
	}

	token, err := jwt.Decode(idToken)
	if err != nil {
		return "", "", err
	}
	return token.Uid, token.Gid, nil
}

func Authenticate(params *types.AuthenticationParams) error {
	ctx := context.Background()
	params, err := GetFinalAuthenticationParams(params)
	if err != nil {
		return err
	}
	log.Debugf("Final authentication params: %v", params)
	token, err := runAuthenticationByFlow(ctx, params)
	if err != nil {
		return err
	}
	log.Debug("Authentication process done successfully")
	if params.User == "" {
		return kubeconfig.SetTokenToCurrentUser(params.AuthenticationFlow, token)
	}
	return kubeconfig.SetTokenToUser(params.User, params.AuthenticationFlow, token)
}

func GetFinalAuthenticationParams(cliParams *types.AuthenticationParams) (*types.AuthenticationParams, error) {
	var kubeConfigParams *types.AuthenticationParams
	var err error
	if cliParams.User == "" {
		kubeConfigParams, err = kubeconfig.GetCurrentUserAuthenticationParams()
	} else {
		kubeConfigParams, err = kubeconfig.GetUserAuthenticationParams(cliParams.User)
	}
	if err != nil {
		return nil, err
	}
	log.Debugf("Read user kubeConfig authentication params: %v", kubeConfigParams)
	cliParams = cliParams.MergeAuthenticationParams(kubeConfigParams)
	return cliParams.ValidateAndSetDefaultAuthenticationParams()
}

func runAuthenticationByFlow(ctx context.Context, params *types.AuthenticationParams) (*oauth2.Token, error) {
	switch params.AuthenticationFlow {
	case types.CodePkceBrowser:
		return code_pkce_browser.AuthenticateCodePkceBrowser(ctx, params)
	case types.Auth0PasswordRealm:
		if util.IsBoolPTrue(params.IsAirgapped) {
			return password.AuthenticateKeycloakPassword(ctx, params)
		}
		return password.AuthenticateAuth0PasswordRealm(ctx, params)
	}
	return nil, fmt.Errorf("unidentified authentication method %v", params.AuthenticationFlow)
}
