package commands

import (
	"strings"
	"testing"
)

func TestNoPvcRequested(t *testing.T) {
	args := &submitArgs{
		PersistentVolumes: []string{},
	}
	err := HandleVolumesAndPvc(args)

	if err != nil {
		t.Errorf("Unexpected error %+v", err)
	} else if len(args.PersistentVolumes) != 0 {
		t.Errorf("Expected args.PersistentVolumes to be of size 0, actual: %+v", args.PersistentVolumes)
	}
}

func TestWrongDirectiveLength(t *testing.T) {
	args := &submitArgs{PersistentVolumes: []string{":"}}
	assertWrongDirectiveLengthError(t, args)

	args = &submitArgs{PersistentVolumes: []string{"::::"}}
	assertWrongDirectiveLengthError(t, args)
}

func assertWrongDirectiveLengthError(t *testing.T, args *submitArgs) {
	err := HandleVolumesAndPvc(args)
	if err == nil {
		t.Error("Expected to get 'wrong length of args in directive' error, but received non")
	} else if !strings.Contains(err.Error(), "--pvc directives must be given in the form of") {
		t.Errorf("Unexpected error: '%+v", err)
	}
}

func TestNoContainerMountPathGiven(t *testing.T) {
	args := &submitArgs{PersistentVolumes: []string{":1Gi:"}}
	assertMissingMountPathError(t, args)

	args = &submitArgs{PersistentVolumes: []string{"storage-class:1Gi:"}}
	assertMissingMountPathError(t, args)

	args = &submitArgs{PersistentVolumes: []string{"storage-class:1Gi::"}}
	assertMissingMountPathError(t, args)

	args = &submitArgs{PersistentVolumes: []string{"storage-class:1Gi::ro"}}
	assertMissingMountPathError(t, args)

	args = &submitArgs{PersistentVolumes: []string{":1Gi::"}}
	assertMissingMountPathError(t, args)

	args = &submitArgs{PersistentVolumes: []string{":1Gi::ro"}}
	assertMissingMountPathError(t, args)
}

func assertMissingMountPathError(t *testing.T, args *submitArgs) {
	err := HandleVolumesAndPvc(args)
	if err == nil {
		t.Error("Error expected when not passing container mount path")
	} else if !strings.Contains(err.Error(), "container mount path must be specified") {
		t.Errorf("Unexpected error: '%+v", err)
	}
}

func TestNoCapacityGiven(t *testing.T) {
	args := &submitArgs{PersistentVolumes: []string{"::/path/to"}}
	assertMissingCapacityError(t, args)

	args = &submitArgs{PersistentVolumes: []string{"::/path/to:"}}
	assertMissingCapacityError(t, args)

	args = &submitArgs{PersistentVolumes: []string{"my-storage-class:::"}}
	assertMissingCapacityError(t, args)

	args = &submitArgs{PersistentVolumes: []string{"my-storage-class::/path/to:"}}
	assertMissingCapacityError(t, args)

	args = &submitArgs{PersistentVolumes: []string{"my-storage-class::/path/to:ro"}}
	assertMissingCapacityError(t, args)

	args = &submitArgs{PersistentVolumes: []string{"::/path/to:ro"}}
	assertMissingCapacityError(t, args)

	args = &submitArgs{PersistentVolumes: []string{"::/path/to:ro"}}
	assertMissingCapacityError(t, args)
}

func assertMissingCapacityError(t *testing.T, args *submitArgs) {
	err := HandleVolumesAndPvc(args)
	if err == nil {
		t.Error("Expected to get a 'missing capacity' error")
	} else if !strings.Contains(err.Error(), "persistent volume size must be specified") {
		t.Errorf("Unexpected error message %+v", err)
	}
}

func TestGoodDirectives(t *testing.T) {
	args := &submitArgs{
		Name: "MyJob",
		PersistentVolumes: []string{
			"storage-class:16Gi:/path/to/mount1",
			"storage-class:16Gi:/path/to/mount2:",
			"storage-class:32Gi:/path/to/mount3:ro",
			":2Gi:/path/to/mount4",
			":2Gi:/path/to/mount5:",
			":2Gi:/path/to/mount6:ro",
		},
	}
	err := HandleVolumesAndPvc(args)
	if err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	assertPvDirective(t, "storage-class:16Gi:/path/to/mount1:", args.PersistentVolumes[0])
	assertPvDirective(t, "storage-class:16Gi:/path/to/mount2:", args.PersistentVolumes[1])
	assertPvDirective(t, "storage-class:32Gi:/path/to/mount3:ro", args.PersistentVolumes[2])
	assertPvDirective(t, ":2Gi:/path/to/mount4:", args.PersistentVolumes[3])
	assertPvDirective(t, ":2Gi:/path/to/mount5:", args.PersistentVolumes[4])
	assertPvDirective(t, ":2Gi:/path/to/mount6:ro", args.PersistentVolumes[5])
}

func assertPvDirective(t *testing.T, expected string, actual string) {
	if expected != actual {
		t.Errorf("Expected persistent volume directive '%s' but got '%s'", expected, actual)
	}
}

func TestStorageClassValidation(t *testing.T) {
	args := &submitArgs{PersistentVolumes: []string{"~oops:2Gi:path:ro"}}
	assertValidation(t, args, "A Storage Class name must consist of")

	args = &submitArgs{PersistentVolumes: []string{"/a:2Gi:path:ro"}}
	assertValidation(t, args, "A Storage Class name must consist of")

	args = &submitArgs{PersistentVolumes: []string{":/2Gi:path:ro"}}
	assertValidation(t, args, "Badly formatted resource request")

	args = &submitArgs{PersistentVolumes: []string{":2Gi~:path:ro"}}
	assertValidation(t, args, "Badly formatted resource request")

	args = &submitArgs{PersistentVolumes: []string{":2Gi:path:ra"}}
	assertValidation(t, args, "invalid readonly directive given")
}

func assertValidation(t *testing.T, args *submitArgs, expectedErr string) {
	err := HandleVolumesAndPvc(args)
	if err == nil {
		t.Errorf("Storage Class validation error expected")
	} else if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error message like '%s' but got '%s'", expectedErr, err.Error())
	}
}

