package code_pkce_remote_browser

import (
	"context"
	"fmt"
	"github.com/run-ai/runai-cli/pkg/authentication/flows"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/util/rand"
)

func AuthenticateCodePkceRemoteBrowser(ctx context.Context, authParams *types.AuthenticationParams) (*oauth2.Token, error) {
	log.Debug("Authentication process start with authorization code flow, with PKCE, remote browser mode")
	oauth2Config, err := flows.GetOauth2Config(ctx, authParams)
	if err != nil {
		return nil, err
	}
	remoteAuthenticationUrl := oauth2Config.AuthCodeURL(rand.String(7), oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser: \n\t%v\n", remoteAuthenticationUrl)
	fmt.Printf("Enter verification code: ")
	var code string
	_, err = fmt.Scanln(&code)
	if err != nil {
		return nil, err
	}

	return oauth2Config.Exchange(ctx, code, oauth2.AccessTypeOffline)
}
