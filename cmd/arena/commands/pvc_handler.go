package commands

import (
	"fmt"
	"strings"
)

//input:
//      [0]          [1]  [2]                [3]
// --pvc StorageClass[optional]:Size:ContainerMountPath:AccessMode[optional]
func handlePvc(args *submitRunaiJobArgs) error {
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
		return nil, fmt.Errorf("--pv directives must be given in the form of StorageClass[optional]:Size:ContainerMountPath:AccessMode[optional], if any field is left blank the delimiting ':' must still be passed")
	} else if mountDirectiveParts[1] == "" {
		return nil, fmt.Errorf("persistent volume size must be specified")
	} else if mountDirectiveParts[2] == "" {
		return nil, fmt.Errorf("container mount path must be specified")
	}
	// normalize length since we're dealing with slice indices.
	if len(mountDirectiveParts) == 3 {
		mountDirectiveParts = append(mountDirectiveParts, "")
	}
	return mountDirectiveParts, nil
}