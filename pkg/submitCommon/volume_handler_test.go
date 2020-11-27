package submitCommon

import (
	"runtime/debug"
	"strings"
	"testing"
)

const (
	WrongLengthPvcParamError           = "the --pvc parameter can be given"
	WrongLengthVolumeParamError        = "the -v parameter must be"
	BadlyFormattedResourceRequestError = "Badly formatted resource request"
	MissingContainerMountPathError     = "container mount path must be specified"
	MissingCapacityError               = "persistent volume size must be specified"
	InvalidReadOnlyParamError          = "invalid readonly parameter given:"
	MissingPvcNameError                = "persistent volume claim name must be specified"
)

func TestNoPvcRequested(t *testing.T) {
	args := &SubmitArgs{
		PersistentVolumes: []string{},
	}
	err := handleVolumesAndPvc(args)

	if err != nil {
		t.Errorf("Unexpected error %+v", err)
	} else if len(args.PersistentVolumes) != 0 {
		t.Errorf("Expected args.PersistentVolumes to be of size 0, actual: %+v", args.PersistentVolumes)
	}
}

func TestVolumeParamLength(t *testing.T) {
	// Wrong length
	args := &SubmitArgs{Volumes: []string{""}}
	assertValidation(t, args, WrongLengthVolumeParamError)

	args = &SubmitArgs{Volumes: []string{":::"}}
	assertValidation(t, args, WrongLengthVolumeParamError)

	args = &SubmitArgs{Volumes: []string{"::::"}}
	assertValidation(t, args, WrongLengthVolumeParamError)

	// Correct length
	args = &SubmitArgs{Volumes: []string{":"}}
	assertNoError(t, args)

	args = &SubmitArgs{Volumes: []string{"myPath:yourPath"}}
	assertNoError(t, args)

	args = &SubmitArgs{Volumes: []string{"::"}}
	assertNoError(t, args)

	args = &SubmitArgs{Volumes: []string{"myPath:yourPath:ro"}}
	assertNoError(t, args)
}

func TestPvcParamLength(t *testing.T) {
	// Wrong length
	args := &SubmitArgs{PersistentVolumes: []string{""}}
	assertValidation(t, args, WrongLengthPvcParamError)

	args = &SubmitArgs{PersistentVolumes: []string{"::::"}}
	assertValidation(t, args, WrongLengthPvcParamError)

	args = &SubmitArgs{PersistentVolumes: []string{":::::"}}
	assertValidation(t, args, WrongLengthPvcParamError)

	args = &SubmitArgs{PersistentVolumes: []string{"storage-class:1Gi::ro:notgood"}}
	assertValidation(t, args, WrongLengthPvcParamError)

	// Correct length

	args = &SubmitArgs{PersistentVolumes: []string{":1Gi:/path"}}
	assertNoError(t, args)

	args = &SubmitArgs{PersistentVolumes: []string{"storageclass:1Gi:/path"}}
	assertNoError(t, args)

	args = &SubmitArgs{PersistentVolumes: []string{":1Gi:/path:"}}
	assertNoError(t, args)

	args = &SubmitArgs{PersistentVolumes: []string{":1Gi:/path:ro"}}
	assertNoError(t, args)

	args = &SubmitArgs{PersistentVolumes: []string{"storageclass:1Gi:/path:"}}
	assertNoError(t, args)

	args = &SubmitArgs{PersistentVolumes: []string{"storageclass:1Gi:/path:ro"}}
	assertNoError(t, args)
}

func TestNoContainerMountPathGiven(t *testing.T) {
	args := &SubmitArgs{PersistentVolumes: []string{":1Gi:"}}
	assertValidation(t, args, MissingContainerMountPathError)

	args = &SubmitArgs{PersistentVolumes: []string{"storage-class:1Gi:"}}
	assertValidation(t, args, MissingContainerMountPathError)

	args = &SubmitArgs{PersistentVolumes: []string{"storage-class:1Gi::"}}
	assertValidation(t, args, MissingContainerMountPathError)

	args = &SubmitArgs{PersistentVolumes: []string{"storage-class:1Gi::ro"}}
	assertValidation(t, args, MissingContainerMountPathError)

	args = &SubmitArgs{PersistentVolumes: []string{":1Gi::"}}
	assertValidation(t, args, MissingContainerMountPathError)

	args = &SubmitArgs{PersistentVolumes: []string{":1Gi::ro"}}
	assertValidation(t, args, MissingContainerMountPathError)
}

func TestNoCapacityGiven(t *testing.T) {
	args := &SubmitArgs{PersistentVolumes: []string{"::/path/to"}}
	assertValidation(t, args, MissingCapacityError)

	args = &SubmitArgs{PersistentVolumes: []string{"::/path/to:"}}
	assertValidation(t, args, MissingCapacityError)

	args = &SubmitArgs{PersistentVolumes: []string{"my-storage-class:::"}}
	assertValidation(t, args, MissingCapacityError)

	args = &SubmitArgs{PersistentVolumes: []string{"my-storage-class::/path/to:"}}
	assertValidation(t, args, MissingCapacityError)

	args = &SubmitArgs{PersistentVolumes: []string{"my-storage-class::/path/to:ro"}}
	assertValidation(t, args, MissingCapacityError)

	args = &SubmitArgs{PersistentVolumes: []string{"::/path/to:ro"}}
	assertValidation(t, args, MissingCapacityError)

	args = &SubmitArgs{PersistentVolumes: []string{"::/path/to:ro"}}
	assertValidation(t, args, MissingCapacityError)
}

func TestExistingPvcNoPvcNameGiven(t *testing.T) {
	args := &SubmitArgs{PersistentVolumes: []string{":/path/to"}}
	assertValidation(t, args, MissingPvcNameError)

	// This will be validated as a dynamic provision param because of the empty string at [0] and len() == 3
	args = &SubmitArgs{PersistentVolumes: []string{":/path/to:ro"}}
	assertValidation(t, args, BadlyFormattedResourceRequestError)

	// This will be validated as a dynamic provision param because of the empty string at [0] and len() == 3
	args = &SubmitArgs{PersistentVolumes: []string{"::ro"}}
	assertValidation(t, args, MissingCapacityError)

	args = &SubmitArgs{PersistentVolumes: []string{":"}}
	assertValidation(t, args, MissingPvcNameError)
}

func TestExistingPvcNoMountPathGiven(t *testing.T) {
	args := &SubmitArgs{PersistentVolumes: []string{"my-pvc:"}}
	assertValidation(t, args, MissingContainerMountPathError)

	args = &SubmitArgs{PersistentVolumes: []string{"my-pvc::ro"}}
	assertValidation(t, args, MissingContainerMountPathError)

	args = &SubmitArgs{PersistentVolumes: []string{"my-pvc::"}}
	assertValidation(t, args, MissingContainerMountPathError)
}

func TestExistingPvcBadAccessModeFlag(t *testing.T) {
	args := &SubmitArgs{PersistentVolumes: []string{"my-pvc:/path1:ra"}}
	assertValidation(t, args, InvalidReadOnlyParamError)

	args = &SubmitArgs{PersistentVolumes: []string{"my-pvc:/path1:oops"}}
	assertValidation(t, args, InvalidReadOnlyParamError)

}

func TestGoodParams(t *testing.T) {
	args := &SubmitArgs{
		Name: "MyJob",
		PersistentVolumes: []string{
			// Dynamic provision pvc
			"storage-class:16Gi:/path/to/mount1",
			"storage-class:16Gi:/path/to/mount2:",
			"storage-class:32Gi:/path/to/mount3:ro",
			":2Gi:/path/to/mount4",
			":2Gi:/path/to/mount5:",
			":2Gi:/path/to/mount6:rw",
			//Pre-existing pvc
			"my-pvc:/path/to",
			"my-pvc:/path2:",
			"my-pvc:/path3:ro",
			"my-pvc:/path4:rw",
		},
	}
	err := handleVolumesAndPvc(args)
	if err != nil {
		t.Fatalf("Unexpected error: %+v", err)
	}
	assertPvParam(t, "storage-class:16Gi:/path/to/mount1:", args.PersistentVolumes[0])
	assertPvParam(t, "storage-class:16Gi:/path/to/mount2:", args.PersistentVolumes[1])
	assertPvParam(t, "storage-class:32Gi:/path/to/mount3:ro", args.PersistentVolumes[2])
	assertPvParam(t, ":2Gi:/path/to/mount4:", args.PersistentVolumes[3])
	assertPvParam(t, ":2Gi:/path/to/mount5:", args.PersistentVolumes[4])
	assertPvParam(t, ":2Gi:/path/to/mount6:rw", args.PersistentVolumes[5])
	assertPvParam(t, "my-pvc:/path/to:", args.PersistentVolumes[6])
	assertPvParam(t, "my-pvc:/path2:", args.PersistentVolumes[7])
	assertPvParam(t, "my-pvc:/path3:ro", args.PersistentVolumes[8])
	assertPvParam(t, "my-pvc:/path4:rw", args.PersistentVolumes[9])
}

func assertPvParam(t *testing.T, expected string, actual string) {
	if expected != actual {
		t.Fatalf("Expected pvc params '%s' but got '%s'\n %v", expected, actual, string(debug.Stack()))
	}
}

func TestStorageClassValidation(t *testing.T) {
	args := &SubmitArgs{PersistentVolumes: []string{"~oops:2Gi:path:ro"}}
	assertValidation(t, args, "A Storage Class name must consist of")

	args = &SubmitArgs{PersistentVolumes: []string{"/a:2Gi:path:ro"}}
	assertValidation(t, args, "A Storage Class name must consist of")

	args = &SubmitArgs{PersistentVolumes: []string{"a:2Gi:path:ro"}}
	assertNoError(t, args)
}

func TestResourceRequestValidation(t *testing.T) {
	args := &SubmitArgs{PersistentVolumes: []string{":/2Gi:path:ro"}}
	assertValidation(t, args, BadlyFormattedResourceRequestError)

	args = &SubmitArgs{PersistentVolumes: []string{":2Gi~:path:ro"}}
	assertValidation(t, args, BadlyFormattedResourceRequestError)

	args = &SubmitArgs{PersistentVolumes: []string{":2Gi:path:ro"}}
	assertNoError(t, args)
}

func TestReadOnlyFlagValidation(t *testing.T) {
	// pvc - bad
	args := &SubmitArgs{PersistentVolumes: []string{":2Gi:/path:ra"}}
	assertValidation(t, args, InvalidReadOnlyParamError)

	args = &SubmitArgs{PersistentVolumes: []string{":2Gi:/path:r"}}
	assertValidation(t, args, InvalidReadOnlyParamError)

	args = &SubmitArgs{PersistentVolumes: []string{":2Gi:/path:re"}}
	assertValidation(t, args, InvalidReadOnlyParamError)

	// pvc - good
	args = &SubmitArgs{PersistentVolumes: []string{":2Gi:/path:ro"}}
	assertNoError(t, args)

	args = &SubmitArgs{PersistentVolumes: []string{":2Gi:/path:rw"}}
	assertNoError(t, args)

	// volume - bad
	args = &SubmitArgs{Volumes: []string{"/myPath:/yourPath:rx"}}
	assertValidation(t, args, InvalidReadOnlyParamError)

	args = &SubmitArgs{Volumes: []string{"/myPath:/yourPath:roo"}}
	assertValidation(t, args, InvalidReadOnlyParamError)

	args = &SubmitArgs{Volumes: []string{"/myPath:/yourPath:rD"}}
	assertValidation(t, args, InvalidReadOnlyParamError)

	// volume - good
	args = &SubmitArgs{Volumes: []string{"/myPath:/yourPath:rw"}}
	assertNoError(t, args)

	args = &SubmitArgs{Volumes: []string{"/myPath:/yourPath:ro"}}
	assertNoError(t, args)

}

func assertValidation(t *testing.T, args *SubmitArgs, expectedErr string) {
	err := handleVolumesAndPvc(args)
	if err == nil {
		t.Fatalf("Error containing '%s expected\n %v", expectedErr, string(debug.Stack()))
	} else if !strings.Contains(err.Error(), expectedErr) {
		t.Fatalf("Expected error message like '%s' but got '%s'\n %v", expectedErr, err.Error(), string(debug.Stack()))
	}
}

func assertNoError(t *testing.T, args *SubmitArgs) {
	err := handleVolumesAndPvc(args)
	if err != nil {
		t.Fatalf("Unexpected error: '%+v' %v", err, string(debug.Stack()))
	}
}
