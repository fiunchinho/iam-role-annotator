package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestWhenOptinAndDeploymentWantsToBeAnnotatedThenDeploymentIsAnnotatedAndSaved(t *testing.T) {
	awsAccountID := "12345"
	applicationName := "app-name"
	expectedAnnotationValue := "arn:aws:iam::" + awsAccountID + ":role/" + applicationName

	deployment := NewDeploymentBuilder().Named(applicationName).WithAnnotation(AnnotationToWatch, "true").Build()

	srv := NewOptinIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	newDeployment, _ := srv.Annotate(deployment)

	assert := assert.New(t)
	assert.Contains(newDeployment.Spec.Template.ObjectMeta.Annotations, Kube2IAMAnnotation, "The deployment PodSpec must contain the kube2iam annotation")
	assert.Equal(expectedAnnotationValue, newDeployment.Spec.Template.ObjectMeta.Annotations[Kube2IAMAnnotation], "The kube2iam annotation must contain the right IAM Role")
}

func TestWhenOptinAndAnnotationIsNotPresentThenDeploymentIsNotAnnotated(t *testing.T) {
	awsAccountID := "12345"
	applicationName := "app-name"

	deployment := NewDeploymentBuilder().Named(applicationName).Build()

	srv := NewOptinIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	newDeployment, _ := srv.Annotate(deployment)

	assert := assert.New(t)
	assert.NotContains(newDeployment.Spec.Template.ObjectMeta.Annotations, Kube2IAMAnnotation, "The deployment PodSpec must not contain the kube2iam annotation")
}

func TestWhenOptinAndDeploymentDoesntWantToBeAnnotatedThenNothingHappens(t *testing.T) {
	awsAccountID := "12345"
	applicationName := "app-name"

	deployment := NewDeploymentBuilder().Named(applicationName).WithAnnotation(AnnotationToWatch, "false").Build()

	srv := NewOptinIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	newDeployment, _ := srv.Annotate(deployment)

	assert := assert.New(t)
	assert.NotContains(newDeployment.Spec.Template.ObjectMeta.Annotations, Kube2IAMAnnotation, "The deployment PodSpec must not contain the kube2iam annotation")
}

func TestWhenOptinDeploymentIsSkippedBecauseAlreadyHasIAMAnnotation(t *testing.T) {
	awsAccountID := "12345"
	applicationName := "app-name"

	roleArn := "arn:aws:iam::" + awsAccountID + ":role/" + applicationName
	deployment := NewDeploymentBuilder().Named(applicationName).WithAnnotation(AnnotationToWatch, "true").WithPodSpecAnnotation(Kube2IAMAnnotation, "previousIAMRoleArn").Build()

	srv := NewOptinIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	deploy, _ := srv.Annotate(deployment)

	assert := assert.New(t)
	assert.Equal(roleArn, deploy.Spec.Template.ObjectMeta.Annotations[Kube2IAMAnnotation])
}

func TestWhenOptoutAndAnnotationIsNotPresentThenDeploymentIsAnnotated(t *testing.T) {
	awsAccountID := "12345"
	applicationName := "app-name"
	expectedAnnotationValue := "arn:aws:iam::" + awsAccountID + ":role/" + applicationName

	deployment := NewDeploymentBuilder().Named(applicationName).Build()

	srv := NewOptoutIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	newDeployment, _ := srv.Annotate(deployment)

	assert := assert.New(t)
	assert.Contains(newDeployment.Spec.Template.ObjectMeta.Annotations, Kube2IAMAnnotation, "The deployment PodSpec must contain the kube2iam annotation")
	assert.Equal(expectedAnnotationValue, newDeployment.Spec.Template.ObjectMeta.Annotations[Kube2IAMAnnotation], "The kube2iam annotation must contain the right IAM Role")
}

func TestWhenOptoutThenDeploymentIsAnnotatedAndSaved(t *testing.T) {
	awsAccountID := "12345"
	applicationName := "app-name"
	expectedAnnotationValue := "arn:aws:iam::" + awsAccountID + ":role/" + applicationName

	deployment := NewDeploymentBuilder().Named(applicationName).Build()

	srv := NewOptoutIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	newDeployment, _ := srv.Annotate(deployment)

	assert := assert.New(t)
	assert.Contains(newDeployment.Spec.Template.ObjectMeta.Annotations, Kube2IAMAnnotation, "The deployment PodSpec must contain the kube2iam annotation")
	assert.Equal(expectedAnnotationValue, newDeployment.Spec.Template.ObjectMeta.Annotations[Kube2IAMAnnotation], "The kube2iam annotation must contain the right IAM Role")
}

func TestWhenOptoutAndAnnotationIsFalseThenDeploymentIsNotAnnotated(t *testing.T) {
	awsAccountID := "12345"
	applicationName := "app-name"

	deployment := NewDeploymentBuilder().Named(applicationName).WithAnnotation(AnnotationToWatch, "false").Build()

	srv := NewOptoutIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	newDeployment, _ := srv.Annotate(deployment)

	assert := assert.New(t)
	assert.NotContains(newDeployment.Spec.Template.ObjectMeta.Annotations, Kube2IAMAnnotation, "The deployment PodSpec must not contain the kube2iam annotation")
}

func NewDeploymentBuilder() *DeploymentBuilder {
	return &DeploymentBuilder{
		annotations: make(map[string]string),
	}
}

type DeploymentBuilder struct {
	name               string
	annotations        map[string]string
	podSpecAnnotations map[string]string
}

func (builder *DeploymentBuilder) Named(name string) *DeploymentBuilder {
	builder.name = name

	return builder
}

func (builder *DeploymentBuilder) WithAnnotation(key string, value string) *DeploymentBuilder {
	builder.annotations[key] = value

	return builder
}

func (builder *DeploymentBuilder) WithPodSpecAnnotation(key string, value string) *DeploymentBuilder {
	builder.podSpecAnnotations[key] = value

	return builder
}

func (builder *DeploymentBuilder) Build() appsv1beta1.Deployment {
	deployment := appsv1beta1.Deployment{}
	deployment.Name = builder.name
	deployment.ObjectMeta.Annotations = builder.annotations
	deployment.Spec.Template.ObjectMeta.Annotations = builder.podSpecAnnotations

	return deployment
}
