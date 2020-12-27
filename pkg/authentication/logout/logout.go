package logout

import (
	"context"
	"fmt"
	"github.com/pkg/browser"
	"github.com/run-ai/runai-cli/pkg/authentication/kubeconfig"
	"github.com/run-ai/runai-cli/pkg/authentication/pages"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
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

	var params *types.AuthenticationParams
	if user == "" {
		params, err = kubeconfig.GetCurrentUserAuthenticationParams()
	} else {
		params, err = kubeconfig.GetUserAuthenticationParams(user)
	}
	if err != nil {
		return err
	}
	params, err = params.ValidateAndSetDefaultAuthenticationParams()
	if err != nil {
		return err
	}

	return logoutUserSSOCookie(params)
	switch params.AuthenticationFlow {
	case types.CodePkceBrowser:
		err = logoutUserSSOCookie(params)
	}
	return err
}

func logoutUserSSOCookie(params *types.AuthenticationParams) error {
	var eg errgroup.Group
	eg.Go(func() error { return serverLogoutWeb(params.ListenAddress) })
	redirectUrl := fmt.Sprintf("%vlogout", params.GetRedirectUrl())
	eg.Go(func() error {
		return browser.OpenURL(fmt.Sprintf("%vv2/logout?returnTo=%v&client_id=%v", params.IssuerURL, url.QueryEscape(redirectUrl), params.ClientId))
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
	_ = s.ListenAndServe()
	return nil
}
