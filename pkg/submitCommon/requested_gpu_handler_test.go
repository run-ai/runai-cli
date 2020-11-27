package submitCommon

import (
	"strconv"
	"testing"
)

type testArgs struct {
	interactive *bool
	gpu         *float64
}

func TestGPUSharingManager(t *testing.T) {
	interactiveTrue := true
	fractionalGPU := 0.2
	wholeGPU := float64(1)

	tests := []struct {
		name                      string
		shouldRunFractionalGPUJob bool
		args                      *testArgs
	}{
		{
			name: "Valid fractional GPU job",
			args: &testArgs{
				interactive: &interactiveTrue,
				gpu:         &fractionalGPU,
			},
			shouldRunFractionalGPUJob: true,
		},
		{
			name: "Valid whole GPU job",
			args: &testArgs{
				interactive: &interactiveTrue,
				gpu:         &wholeGPU,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			submitArgs := setSubmitArgs(tt.args)
			testSubmitArgs := *submitArgs
			handleRequestedGPUs(&testSubmitArgs)

			gpuFraction, err := strconv.ParseFloat(testSubmitArgs.GPUFraction, 64)
			if err != nil {
				if tt.shouldRunFractionalGPUJob {
					t.Errorf("handleSharedGPUsIfNeeded() failed to parse gpuFraction %v, while expecting it to manage", err)
				} else if float64(*testSubmitArgs.GPUInt) != *submitArgs.GPU {
					t.Errorf("GPUInt: %v, SubmitArgs.gpu: %v", *testSubmitArgs.GPUInt, *submitArgs.GPU)
				}
			}

			if gpuFraction != *submitArgs.GPU && tt.shouldRunFractionalGPUJob {
				t.Errorf("gpuFraction: %v, *testSubmitArgs.gpu: %v, miss match", gpuFraction, *testSubmitArgs.GPU)
			}
		})
	}
}

func setSubmitArgs(args *testArgs) *SubmitArgs {
	submitArgs := SubmitArgs{}
	submitArgs.GPU = args.gpu
	submitArgs.Interactive = args.interactive
	return &submitArgs
}
