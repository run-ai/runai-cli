package commands

import (
	"fmt"
	validate "github.com/kubeflow/arena/pkg/util"
	"strings"
)

// Input: (surrounded by [] means optional)
//
//            [0]          [1]           [2]                [3]
// --pvc [<storage_class>]:<size>:<container_mount_path>:[<access_mode>]
//
// Or
//
//            [0]					[1]					[2]
// --pvc <existing_pvc_name>:<container_mount_path>:[<access_mode>]
func HandleVolumesAndPvc(args *submitArgs) error {
	if err := handlePvcParams(args); err != nil {
		return err
	} else if err := handleVolumes(args); err != nil {
		return err
	}
	return nil
}

func handlePvcParams(args *submitArgs) (err error) {
	var rebuiltPersistentVolumeParams []string
	for _, joinedPvcParams := range args.PersistentVolumes {
		pvcParams := strings.Split(joinedPvcParams, ":")
		if len(pvcParams) < 2 || len(pvcParams) > 4 {
			return fmt.Errorf("the --pvc parameter can be given in either of the forms: <storage_class>[optional]:<size>:<container_mount_path>:<access_mode>[optional] " +
				"or <existing_pvc_name>:<container_mount_path>:<access_mode>[optional] \n If any field is left blank the delimiting ':' must still be passed")
		}
		var rebuiltParams string
		handleAsExistingPvc := extrapolateParamType(pvcParams)
		if handleAsExistingPvc {
			rebuiltParams, err = handleExistingPvc(pvcParams)
			rebuiltPersistentVolumeParams = append(rebuiltPersistentVolumeParams, rebuiltParams)

		} else {
			rebuiltParams, err = handleDynamicProvisionPvc(pvcParams)
			rebuiltPersistentVolumeParams = append(rebuiltPersistentVolumeParams, rebuiltParams)
		}
		if err != nil {
			break
		}
	}

	// Doesn't really matter what we put here if err is not nil
	args.PersistentVolumes = rebuiltPersistentVolumeParams
	return err
}

func extrapolateParamType(pvcParams []string) (handleAsExistingPvc bool) {
	if len(pvcParams) == 2 {
		//len == 2 can either be 'existing-pvc:/path' or error
		handleAsExistingPvc = true
	} else if len(pvcParams) == 4 {
		// len == 4 can either be 'storageClass:2Gi:/path:ro', ':2Gi:/path:ro' or error
		handleAsExistingPvc = false
	} else {
		//len == 3 is the tricky part since both parameter forms can have length 3. for instance ':2Gi:/path' and 'my-pvc:/path:ro'
		// In this case 2 checks can be made to tell us which is which:
		if pvcParams[0] == "" {
			// If params[0] == "" then we treat as dynamic provision form since only there we allow the first part to be empty (in order to use the default storage class)
			handleAsExistingPvc = false
		} else if validationErr := validate.ValidateStorageResourceRequest(pvcParams[1]); validationErr == nil {
			// If params[1] is formatted like a resource request then we have to treat as dynamic provision
			handleAsExistingPvc = false
		} else if pvcParams[2] == validate.AccessModeReadOnlyParam || pvcParams[2] == validate.AccessModeReadWriteParam {
			// If params[2] == ro || params[2] == rw   we';; treat as existing pvc (since the params are of form '?:?:rw/ro'
			handleAsExistingPvc = true
		} else {
			// Otherwise, we have a param string of the form '?:?:?' and we know:
			// 		[0] != "",
			//		[1] is not a resource request,
			//		[2] != ro/rw
			//
			// So it can be either:
			//   - an erroneous dynamic-pvc param, like 'storage-class:badly-formatted-2g:/path'
			//   - an existing-pvc param, like 'pvc-name:/path:'
			//
			// The safest bet is to try it like an existing pvc.
			handleAsExistingPvc = true
		}
	}
	return handleAsExistingPvc
}

// Assumption: len(pvcParams) == 2 || len(pvcParams) == 3
func handleExistingPvc(pvcParams []string) (rebuiltParams string, err error) {
	if len(pvcParams) == 2 {
		// Always normalize to len() == 3 to make it easier for us in the helm chart
		pvcParams = append(pvcParams, "")
	}
	if pvcParams[0] == "" {
		err = fmt.Errorf("persistent volume claim name must be specified")
	} else if pvcParams[1] == "" {
		err = fmt.Errorf("container mount path must be specified")
	} else {
		err = validate.ValidateMountReadOnlyFlag(pvcParams[2])
	}
	return fmt.Sprintf("%s:%s:%s", pvcParams[0], pvcParams[1], pvcParams[2]), err
}

// Assumption: len(pvcParams) == 3 || len(pvcParams) == 4
func handleDynamicProvisionPvc(pvcParams []string) (rebuiltParams string, err error) {
	if len(pvcParams) == 3 {
		// Always normalize to len() == 4 to make it easier for us in the helm chart
		pvcParams = append(pvcParams, "")
	}
	if pvcParams[1] == "" {
		err = fmt.Errorf("persistent volume size must be specified")
	} else if pvcParams[2] == "" {
		err = fmt.Errorf("container mount path must be specified")
	} else if err = validate.ValidateStorageClassName(pvcParams[0]); err != nil {
		// will be returned below
	} else if err = validate.ValidateStorageResourceRequest(pvcParams[1]); err != nil {
		// will be returned below
	} else if err = validate.ValidateMountReadOnlyFlag(pvcParams[3]); err != nil {
		// will be returned below
	}
	return fmt.Sprintf("%s:%s:%s:%s", pvcParams[0], pvcParams[1], pvcParams[2], pvcParams[3]), err
}

func handleVolumes(args *submitArgs) error {
	for _, joinedVolumeParams := range args.Volumes {
		volumeParams := strings.Split(joinedVolumeParams, ":")
		if len(volumeParams) != 2 && len(volumeParams) != 3 {
			return fmt.Errorf("the -v parameter must be given in the form of <host_path>:<container_path>:<access_mode>[optional]")
		}
		if len(volumeParams) == 3 && volumeParams[2] != "" {
			return validate.ValidateMountReadOnlyFlag(volumeParams[2])
		}
	}
	return nil
}
