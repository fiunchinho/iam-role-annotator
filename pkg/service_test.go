package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeploymentIsAnnotatedAndSaved(t *testing.T) {
	awsAccountID := "12345"
	applicationName := "app-name"
	expectedAnnotationValue := "arn:aws:iam::" + awsAccountID + ":role/" + applicationName

	deployment := NewDeploymentBuilder().Named(applicationName).WithAnnotation(AnnotationToWatch, "true").Build()

	srv := NewIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	newDeployment, _ := srv.Annotate(deployment)

	assert := assert.New(t)
	assert.Contains(newDeployment.Spec.Template.ObjectMeta.Annotations, Kube2IAMAnnotation, "The deployment PodSpec must contain the kube2iam annotation")
	assert.Equal(expectedAnnotationValue, newDeployment.Spec.Template.ObjectMeta.Annotations[Kube2IAMAnnotation], "The kube2iam annotation must contain the right IAM Role")
}

func TestDeploymentDoesntWantToBeAnnotatedThenNothingHappens(t *testing.T) {
	awsAccountID := "12345"
	applicationName := "app-name"

	deployment := NewDeploymentBuilder().Named(applicationName).Build()

	srv := NewIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	newDeployment, _ := srv.Annotate(deployment)

	assert := assert.New(t)
	assert.NotContains(newDeployment.Spec.Template.ObjectMeta.Annotations, Kube2IAMAnnotation, "The deployment PodSpec must not contain the kube2iam annotation")
}

func TestDeploymentIsSkippedBecauseAlreadyHasKube2IAMAnnotation(t *testing.T) {
	awsAccountID := "12345"
	applicationName := "app-name"

	roleArn := "arn:aws:iam::" + awsAccountID + ":role/" + applicationName
	deployment := NewDeploymentBuilder().Named(applicationName).WithAnnotation(AnnotationToWatch, "true").WithAnnotation(Kube2IAMAnnotation, roleArn).Build()

	srv := NewIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	srv.Annotate(deployment)
}

func NewDeploymentBuilder() *DeploymentBuilder {
	return &DeploymentBuilder{
		annotations: make(map[string]string),
	}
}

type DeploymentBuilder struct {
	name        string
	annotations map[string]string
}

func (builder *DeploymentBuilder) Named(name string) *DeploymentBuilder {
	builder.name = name

	return builder
}

func (builder *DeploymentBuilder) WithAnnotation(key string, value string) *DeploymentBuilder {
	builder.annotations[key] = value

	return builder
}

func (builder *DeploymentBuilder) Build() appsv1beta1.Deployment {
	deployment := appsv1beta1.Deployment{}
	deployment.Name = builder.name
	deployment.ObjectMeta.Annotations = builder.annotations

	return deployment
}
