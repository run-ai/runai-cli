package authentication

import (
	"context"
	"fmt"
	"github.com/run-ai/runai-cli/pkg/authentication/kubeconfig"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	openIdScope       = "openid"
	refreshTokenScope = "offline_access"
)

func Authenticate(params *types.AuthenticationParams) error {
	ctx := context.Background()
	var kubeConfigParams *types.AuthenticationParams
	var err error
	if params.User == "" {
		kubeConfigParams, err = kubeconfig.GetCurrentUserAuthenticationParams()
	} else {
		kubeConfigParams, err = kubeconfig.GetUserAuthenticationParams(params.User)
	}
	if err != nil {
		return err
	}
	log.Debugf("Read user kubeConfig authentication params: %v", kubeConfigParams)
	params = params.MergeAuthenticationParams(kubeConfigParams)
	params, err = params.ValidateAndSetDefaultAuthenticationParams()
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

func runAuthenticationByFlow(ctx context.Context, params *types.AuthenticationParams) (*oauth2.Token, error) {
	switch params.AuthenticationFlow {
	case types.CodePkceBrowser:
		return authenticateCodePkceBrowser(ctx, params)
	}
	return nil, fmt.Errorf("unidentified authentication methd %v", params.AuthenticationFlow)
}
