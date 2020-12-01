package auth

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/client"
	authv1 "k8s.io/api/authorization/v1"
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
		err = GetKubeLoginErrorIfNeeded(apiErr)
	} else if !permissionResponse.Status.Allowed {
		err = getKubeLoginError(fmt.Errorf("permission Check in API Server returned false"))
	}
	return err
}
