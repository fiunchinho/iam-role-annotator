package pkg

import (
	"testing"

	appsv1beta1 "k8s.io/api/apps/v1beta1"

	"k8s.io/client-go/kubernetes/fake"
)

func TestRetrieveDeployment(t *testing.T) {
	deploymentRetrieve := NewDeploymentRetrieve("namespace", fake.NewSimpleClientset())

	deployment := deploymentRetrieve.GetObject()
	_, ok := deployment.(*appsv1beta1.Deployment)
	if !ok {
		t.Error("Should retrieve Deployment objects")
	}
}
