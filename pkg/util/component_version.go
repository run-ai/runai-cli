package util

import (
	"context"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/clusterConfig"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckComponentVersion checks the version of a runai component
func CheckComponentVersion(componentName string, constraint string, kubeClient *client.Client) bool {
	componentDeployment, err := kubeClient.GetClientset().AppsV1().Deployments(
		clusterConfig.RunaiNamespace).Get(context.TODO(), componentName, metav1.GetOptions{})
	if err != nil {
		log.Debugf("Failed to query version of %s\n", componentName)
		return true
	}
	// assuming the first container has the main component image
	componentImageParts := strings.Split(componentDeployment.Spec.Template.Spec.Containers[0].Image, ":")
	componentVersion, err := version.NewSemver(componentImageParts[len(componentImageParts)-1])
	if err != nil {
		log.Debugf("Failed to parse version of %s: %s\n", componentName, componentImageParts[len(componentImageParts)-1])
		return true
	}
	givenConstraint, err := version.NewConstraint(constraint)
	if err != nil {
		log.Debugf("Failed to parse constraint for %s: %s\n", componentName, constraint)
		return true
	}
	return givenConstraint.Check(componentVersion)
}
