package submit

import "fmt"

func handleImagePullPolicy(submitArgs *submitArgs) (err error) {
	switch submitArgs.ImagePullPolicy {
	case "always", "ifNotPresent", "never":
		if submitArgs.LocalImage != nil && *submitArgs.LocalImage {
			submitArgs.ImagePullPolicy = "never"
		}
		if submitArgs.AlwaysPullImage != nil && *submitArgs.AlwaysPullImage {
			submitArgs.ImagePullPolicy = "always"
		}
		return nil
	default:
		return fmt.Errorf("--imagePullPolicy should be one of: always, ifNotPresent or never")
	}
}
