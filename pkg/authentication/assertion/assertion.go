package assertion

import (
	"context"
	"fmt"
	"github.com/run-ai/runai-cli/pkg/client"
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func AssertViewerRole() error {
	return assertPermission(authv1.SelfSubjectAccessReviewSpec{
		ResourceAttributes: &authv1.ResourceAttributes{
			Verb:     "list",
			Group:    "run.ai",
			Version:  "v1",
			Resource: "projects",
		},
	})
}

func AssertExecutorRole(namespace string) error {
	return assertPermission(authv1.SelfSubjectAccessReviewSpec{
		ResourceAttributes: &authv1.ResourceAttributes{
			Verb:      "create",
			Group:     "",
			Version:   "v1",
			Resource:  "configmaps",
			Namespace: namespace,
		},
	})
}

func assertPermission(request authv1.SelfSubjectAccessReviewSpec) (err error) {
	kubeClient, err := client.GetClient()
	if err != nil {
		return err
	}
	permissionResponse, err := kubeClient.GetClientset().AuthorizationV1().SelfSubjectAccessReviews().Create(
		context.TODO(), &authv1.SelfSubjectAccessReview{Spec: request}, metav1.CreateOptions{})
	if err != nil {
		err = getAuthorizationErrorIfNeeded(err)
	} else if !permissionResponse.Status.Allowed {
		err = getUnauthorizedError()
	}
	return err
}

func getAuthorizationErrorIfNeeded(inputErr error) error {
	if isNoValidTokenExists(inputErr) {
		return fmt.Errorf("User not authenticated, run the ‘runai login’ command.")
	} else if errors.IsForbidden(inputErr) || errors.IsUnauthorized(inputErr) {
		return getUnauthorizedError()
	}
	return inputErr
}

func isNoValidTokenExists(inputErr error) bool {
	return strings.Contains(fmt.Sprintf("%s", inputErr), "No valid id-token")
}

func getUnauthorizedError() error {
	return fmt.Errorf("Access denied. You are not authorized to perform this action.")
}
