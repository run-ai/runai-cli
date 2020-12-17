package auth

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/auth/config"
	"github.com/run-ai/runai-cli/pkg/auth/jwt"
	"github.com/run-ai/runai-cli/pkg/auth/util"
	"github.com/run-ai/runai-cli/pkg/client"
	log "github.com/sirupsen/logrus"
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// This is a naive assertion which checks just one action the viewer should be allowed to do - the full permission range can be found in installkit
func AssertViewerRole() error {
	return AssertPermission(authv1.SelfSubjectAccessReviewSpec{
		ResourceAttributes: &authv1.ResourceAttributes{
			Verb:     "list",
			Group:    "run.ai",
			Version:  "v1",
			Resource: "projects",
		},
	})
}

// This is a naive assertion which checks just one action the executor should be allowed to do - the full permission range can be found in installkit
func AssertExecutorRole(namespace string) error {
	return AssertPermission(authv1.SelfSubjectAccessReviewSpec{
		ResourceAttributes: &authv1.ResourceAttributes{
			Verb:     "create",
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
			Namespace: namespace,
		},
	})
}

func AssertPermission(request authv1.SelfSubjectAccessReviewSpec) (err error) {
	kubeClient, err := client.GetClient()
	if err != nil {
		return err
	}
	if permissionResponse, apiErr := kubeClient.GetClientset().AuthorizationV1().SelfSubjectAccessReviews().Create(&authv1.SelfSubjectAccessReview{Spec: request}); apiErr != nil {
		err = GetPermissionErrorIfNeeded(apiErr)
	} else if !permissionResponse.Status.Allowed {
		err = getPermissionError(fmt.Errorf("permission Check in API Server returned false"))
	}
	return err
}


func GetPermissionErrorIfNeeded(err error) error {
	if isAuthError(err) {
		err = getPermissionError(err)
	}
	return err
}

func getPermissionError(err error) error {
	if username, permissionErr := GetEmailForCurrentUser(); permissionErr != nil {
		log.Debug("Can't acquire username from current user's token: ", permissionErr)
	} else if username != "" {
		//Write the original message to debug log so we can actually understand what's going on.
		log.Debug(err)
		err = fmt.Errorf("user %s doesn't have the required permissions to perform this operation", username)
	}
	return err
}

func isAuthError(err error) bool {
	return errors.IsForbidden(err) || errors.IsUnauthorized(err)
}

// This logic:
// Gets the current kubeconfig context
// Gets the user from the current context
// Gets auth provider config from user
// Gets ID token from auth provider config
// Decodes ID token and returns the email scope.
func GetEmailForCurrentUser() (email string, err error) {
	var token jwt.Token
	var rawIdToken string
	if rawIdToken, err = getTokenForCurrentKubectlUser(); err == nil {
		if token, err = jwt.Decode(rawIdToken); err == nil {
			email = token.Email
		}
	}
	return email, err
}

func getTokenForCurrentKubectlUser() (rawIdToken string, err error) {
	kubeConfig, err := util.ReadKubeConfig()
	if err != nil {
		return rawIdToken, fmt.Errorf("failed to parse kubeconfig file: %v", err)
	}
	currentKubeConfigUser := kubeConfig.Contexts[kubeConfig.CurrentContext].AuthInfo
	userAuth, ok := kubeConfig.AuthInfos[currentKubeConfigUser]
	if !ok {
		return rawIdToken, fmt.Errorf("No auth configuration found in kubeconfig for user '%s' ", currentKubeConfigUser)
	}
	if userAuth.AuthProvider != nil && len(userAuth.AuthProvider.Config) > 0 {
		return userAuth.AuthProvider.Config[config.ParamIdToken], nil
	}
	return rawIdToken, fmt.Errorf("No auth configuration found in kubeconfig for user '%s' ", currentKubeConfigUser)
}