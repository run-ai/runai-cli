package util

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NAMESPACE = "namespace"
)

var ()

func TestComponentVersionCheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Check components versions")
}

type testData struct {
	imageName           string
	constraint          string
	deploymentName      string
	testedComponentName string
	expectedResult      bool
}

var _ = Describe("CheckComponentVersion", func() {
	tests := map[string]testData{
		"returns true if deployment not found": {
			imageName:           "hello:world",
			deploymentName:      "hello",
			testedComponentName: "world",
			expectedResult:      true,
		},
		"image tag has no version": {
			imageName:           "hello:world",
			deploymentName:      "my-controller",
			testedComponentName: "my-controller",
			constraint:          ">=1.2.2",
			expectedResult:      true,
		},
		"fails if image tag has too low version": {
			imageName:           "hello:1.2.1",
			deploymentName:      "my-controller",
			testedComponentName: "my-controller",
			constraint:          ">=1.2.2",
			expectedResult:      false,
		},
		"success if image tag has great enough version": {
			imageName:           "hello:1.2.3",
			deploymentName:      "my-controller",
			testedComponentName: "my-controller",
			constraint:          ">=1.2.2",
			expectedResult:      true,
		},
		"handles versions that start with v": {
			imageName:           "hello:v1.2.1",
			deploymentName:      "my-controller",
			testedComponentName: "my-controller",
			constraint:          ">=v1.2.2",
			expectedResult:      false,
		},
	}
	for testName, data := range tests {
		testName := testName
		td := data
		It(testName, func() {
			deployment := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      td.deploymentName,
					Namespace: "runai",
				},
				Spec: appsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: td.imageName,
								},
							},
						},
					},
				},
			}
			objects := []runtime.Object{&deployment}
			kubeClient, _ := GetClientWithObject(objects)
			result := CheckComponentVersion(td.testedComponentName, td.constraint, &kubeClient)
			Expect(result).To(Equal(td.expectedResult))
		})
	}
})
