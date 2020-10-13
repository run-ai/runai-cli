package submit

import (
	"testing"
)

type pullPolicyTestArgs struct {
	imagePullPolicy string
	alwaysPullImage *bool
	localImage *bool
}

func TestHandleImagePullPolicy(t *testing.T) {
	tests := []struct {
		name                      string
		expectedPullPolicy		  string
		args                      *pullPolicyTestArgs
	}{
		{
			name: "valid ImagePullPolicy",
			args: &pullPolicyTestArgs{
				imagePullPolicy: "always",
				alwaysPullImage:         nil,
				localImage: nil,
			},
			expectedPullPolicy: "always",
		},
		{
			name: "invalid ImagePullPolicy",
			args: &pullPolicyTestArgs{
				imagePullPolicy: "invalid value",
				alwaysPullImage:         nil,
				localImage: nil,
			},
			expectedPullPolicy: "", // expected error
		},
		{
			name: "localImageIsTrue is true",
			args: &pullPolicyTestArgs{
				imagePullPolicy: "always",
				alwaysPullImage:         nil,
				localImage: &[]bool{true}[0],
			},
			expectedPullPolicy: "never",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			submitArgs := setPullPolicySubmitArgs(tt.args)
			testSubmitArgs := *submitArgs
			err := handleImagePullPolicy(&testSubmitArgs)
			pullPolicy := testSubmitArgs.ImagePullPolicy
			if err != nil && tt.expectedPullPolicy != "" {
				t.Errorf("handleImagePullPolicy failed with error. imagePullPolicy: %s, alwaysPullImage: %v, localImage: %v", tt.args.imagePullPolicy, tt.args.alwaysPullImage, tt.args.localImage)
			} else if pullPolicy != tt.expectedPullPolicy {
				t.Errorf("unexpected pull pollicy value. expectedPullPolicy: %s, imagePullPolicy: %s, alwaysPullImage: %v, localImage: %v", tt.expectedPullPolicy, tt.args.imagePullPolicy, tt.args.alwaysPullImage, tt.args.localImage)
			}
		})
	}
}

func setPullPolicySubmitArgs(args *pullPolicyTestArgs) *submitArgs {
	submitArgs := submitArgs{}
	submitArgs.ImagePullPolicy = args.imagePullPolicy
	submitArgs.AlwaysPullImage = args.alwaysPullImage
	submitArgs.LocalImage = args.localImage
	return &submitArgs
}