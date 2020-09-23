// Copyright 2019 The Kubeflow Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_RunaiJob(runaijob *RunaiJob) {
	// set default BackoffLimit
	if runaijob.Spec.Completions == nil && runaijob.Spec.Parallelism == nil {
		runaijob.Spec.Completions = new(int32)
		*runaijob.Spec.Completions = 1
		runaijob.Spec.Parallelism = new(int32)
		*runaijob.Spec.Parallelism = 1
	}
	if runaijob.Spec.Parallelism == nil {
		runaijob.Spec.Parallelism = new(int32)
		*runaijob.Spec.Parallelism = 1
	}
	if runaijob.Spec.BackoffLimit == nil {
		runaijob.Spec.BackoffLimit = new(int32)
		*runaijob.Spec.BackoffLimit = 6
	}
	labels := runaijob.Spec.Template.Labels
	if labels != nil && len(runaijob.Labels) == 0 {
		runaijob.Labels = labels
	}
}
