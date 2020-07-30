package commands

import (
	"fmt"
	validate "github.com/kubeflow/arena/pkg/util"
	"strings"
)

//input:
//      [0]          [1]  [2]                [3]
// --pvc StorageClass[optional]:Size:ContainerMountPath:AccessMode[optional]
func HandleVolumesAndPvc(args *submitArgs) error {
	 if err := handlePvc(args); err != nil {
	 	return err
	 } else if err := handleVolumes(args); err != nil {
	 	return err
	 }
	return nil
}

func handlePvc(args *submitArgs) error {
	var rebuiltDirectives []string
	for _, mountDirective := range args.PersistentVolumes {
		mountDirectiveParts, err := splitDirectiveAndValidate(mountDirective)
		if err != nil {
			return err
		}
		rebuiltDirectives = append(rebuiltDirectives, fmt.Sprintf("%s:%s:%s:%s", mountDirectiveParts[0], mountDirectiveParts[1], mountDirectiveParts[2], mountDirectiveParts[3]))
	}
	args.PersistentVolumes = rebuiltDirectives
	return nil
}

func splitDirectiveAndValidate(mountDirective string) ([]string, error) {
	mountDirectiveParts := strings.Split(mountDirective, ":")
	if len(mountDirectiveParts) < 3 || len(mountDirectiveParts) > 4 {
		return nil, fmt.Errorf("--pvc directives must be given in the form of StorageClass[optional]:Size:ContainerMountPath:AccessMode[optional], if any field is left blank the delimiting ':' must still be passed")
	} else if mountDirectiveParts[1] == "" {
		return nil, fmt.Errorf("persistent volume size must be specified")
	} else if mountDirectiveParts[2] == "" {
		return nil, fmt.Errorf("container mount path must be specified")
	}
	// normalize length since we're dealing with slice indices.
	if len(mountDirectiveParts) == 3 {
		mountDirectiveParts = append(mountDirectiveParts, "")
	}
	if err := validate.ValidateStorageClassName(mountDirectiveParts[0]); err != nil {
		return nil, err
	} else if err := validate.ValidateMountReadOnlyFlag(mountDirectiveParts[3]); err != nil {
		return nil, err
	} else if err := validate.ValidateStorageResourceRequest(mountDirectiveParts[1]); err != nil {
		return nil, err
	}
	return mountDirectiveParts, nil
}

func handleVolumes(args *submitArgs) error {
	for _, volumeDirective := range args.Volumes {
		volumeDirectiveParts := strings.Split(volumeDirective, ":")
		if len(volumeDirectiveParts) == 3 && volumeDirectiveParts[2] != "" {
			return validate.ValidateMountReadOnlyFlag(volumeDirectiveParts[2])
		}
	}
	return nil
}