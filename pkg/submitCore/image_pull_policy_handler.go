package submitCore

import (
	"fmt"
)

const (
	pullPolicyAlways       = "Always"
	pullPolicyIfNotPresent = "IfNotPresent"
	pullPolicyNever        = "Never"
)

func handleImagePullPolicy(submitArgs *SubmitArgs) (err error) {
	switch submitArgs.ImagePullPolicy {
	case pullPolicyAlways, pullPolicyIfNotPresent, pullPolicyNever:
		if submitArgs.LocalImage != nil && *submitArgs.LocalImage {
			submitArgs.ImagePullPolicy = pullPolicyNever
		}
		return nil
	default:
		return fmt.Errorf("image-pull-policy flag should be set with one of the following values of: \"Always\", \"IfNotPresent\" or \"Never\"")
	}
}
