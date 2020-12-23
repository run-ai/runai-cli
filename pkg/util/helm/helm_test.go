package helm

import (
	"gotest.tools/assert"
	"testing"
)

func TestGetHelmVersionFromOutputOnlyVersion(t *testing.T) {
	output := "v3.3.4+ga61ce56"
	version, _ := getHelmVersionFromOutput(output)
	assert.Equal(t, version, HELM_3)
}

func TestGetHelmVersionFromOutputWithOutputWarnings(t *testing.T) {
	output := "WARNING: Kubernetes configuration file is group-readable. This is insecure. Location: /home/olegi/.kube/config\nWARNING: Kubernetes configuration file is world-readable. This is insecure. Location: /home/olegi/.kube/config\nv3.3.4+ga61ce56"
	version, _ := getHelmVersionFromOutput(output)
	assert.Equal(t, version, HELM_3)
}

func TestGetHelmVersionFromOutputWithStringNoVersion(t *testing.T) {
	output := "WARNING: Kubernetes configuration file is group-readable. This is insecure. Location: /home/olegi/.kube/config"
	_, err := getHelmVersionFromOutput(output)
	if err == nil {
		t.FailNow()
	}
}

func TestGetHelmVersionFromOutputWithEmptyString(t *testing.T) {
	output := ""
	_, err := getHelmVersionFromOutput(output)
	if err == nil {
		t.FailNow()
	}
}
