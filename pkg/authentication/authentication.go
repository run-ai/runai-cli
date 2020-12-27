package authentication

import (
	"context"
	"fmt"
	"github.com/run-ai/runai-cli/pkg/authentication/kubeconfig"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	"golang.org/x/oauth2"
)

const (
	openIdScope       = "openid"
	refreshTokenScope = "offline_access"
)

func Authenticate(params *types.AuthenticationParams) (*oauth2.Token, error) {
	ctx := context.Background()
	kubeconfigParams, err := kubeconfig.GetCurrentUserAuthenticationParams()
	if err != nil {
		return nil, err
	}
	params = params.MergeAuthenticationParams(kubeconfigParams)
	params, err = params.ValidateAndSetDefaultAuthenticationParams()
	if err != nil {
		return nil, err
	}
	return runAuthenticationByFlow(ctx, params)
}

func runAuthenticationByFlow(ctx context.Context, params *types.AuthenticationParams) (*oauth2.Token, error) {
	switch params.AuthenticationFlow {
	case types.CodePkceBrowser:
		return authenticateCodePkceBrowser(ctx, params)
	}
	return nil, fmt.Errorf("unidentified authentication methd %v", params.AuthenticationFlow)
}
