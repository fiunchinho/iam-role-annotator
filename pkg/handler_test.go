package pkg

import (
	"testing"

	appsv1beta1 "k8s.io/api/apps/v1beta1"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

type IamRoleAnnotatorMock struct {
	client        kubernetes.Interface
	awsAccountID  string
	logger        Logger
	HasBeenCalled bool
}

func (s *IamRoleAnnotatorMock) Annotate(deployment appsv1beta1.Deployment) (*appsv1beta1.Deployment, error) {
	s.HasBeenCalled = true

	return &deployment, nil
}

func TestHandlerWhenAdding(t *testing.T) {
	assert := assert.New(t)
	deployment := NewDeploymentBuilder().Named("my-app").Build()
	iamRoleAnnotator := IamRoleAnnotatorMock{
		client:        fake.NewSimpleClientset(),
		awsAccountID:  "12345",
		logger:        NewDummyLogger(),
		HasBeenCalled: false,
	}
	handler := NewHandler(&iamRoleAnnotator)
	handler.Add(&deployment)
	assert.True(iamRoleAnnotator.HasBeenCalled)

	assert.Nil(handler.Delete("some obj"), "It should do nothing when deleting Deployment objects")
}

func TestHandlerWhenDeleting(t *testing.T) {
	assert := assert.New(t)
	iamRoleAnnotator := IamRoleAnnotatorMock{
		client:        fake.NewSimpleClientset(),
		awsAccountID:  "12345",
		logger:        NewDummyLogger(),
		HasBeenCalled: false,
	}
	handler := NewHandler(&iamRoleAnnotator)

	assert.Nil(handler.Delete("some obj"), "It should do nothing when deleting Deployment objects")
}
