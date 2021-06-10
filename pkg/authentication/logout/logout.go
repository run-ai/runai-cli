package logout

import (
	"context"
	"fmt"
	"github.com/pkg/browser"
	"github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/authentication"
	"github.com/run-ai/runai-cli/pkg/authentication/kubeconfig"
	"github.com/run-ai/runai-cli/pkg/authentication/pages"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"net/http"
	"net/url"
)

func Logout(user string) error {
	var err error
	if user == "" {
		err = kubeconfig.DeleteTokenToCurrentUser()
	} else {
		err = kubeconfig.DeleteTokenToUser(user)
	}
	if err != nil {
		return err
	}
	log.Debug("Tokens deleted")

	var emptyParams types.AuthenticationParams
	params, err := authentication.CalculateAuthenticationParams(&emptyParams)
	if err != nil {
		return err
	}
	log.Debugf("Final authentication params: %v", params)

	switch params.AuthenticationFlow {
	case types.CodePkceBrowser:
		err = logoutUserSSOCookie(params)
	}
	log.Debug("Logout process done successfully")
	return err
}

func logoutUserSSOCookie(params *types.AuthenticationParams) error {
	log.Debug("Clear browser cache cookies")
	var eg errgroup.Group
	eg.Go(func() error { return serverLogoutWeb(params.ListenAddress) })
	eg.Go(func() error {
		redirectUrl := fmt.Sprintf("%vlogout", params.GetRedirectUrl())
		logoutUrl := getSSOLogoutUrl(util.IsBoolPTrue(params.IsAirgapped), params.IssuerURL, redirectUrl, params.ClientId)
		log.Debugf("Open browser url: %v", logoutUrl)
		return browser.OpenURL(logoutUrl)
	})

	return eg.Wait()
}

func serverLogoutWeb(server string) error {
	s := http.Server{Addr: server, Handler: nil}
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		logoutPage := pages.LogoutPageHtml
		fmt.Fprintf(w, logoutPage)
		go s.Shutdown(context.TODO())
	})
	log.Debug("Open server to redirect after browser logout")
	_ = s.ListenAndServe()
	return nil
}

func getSSOLogoutUrl(isAirgapped bool, issuerURL, redirectUrl, clientId string) string {
	if isAirgapped {
		return fmt.Sprintf("%v/protocol/openid-connect/logout?redirect_uri=%v", issuerURL, redirectUrl)
	}
	return fmt.Sprintf("%vv2/logout?returnTo=%v&client_id=%v", issuerURL, url.QueryEscape(redirectUrl), clientId)
}
