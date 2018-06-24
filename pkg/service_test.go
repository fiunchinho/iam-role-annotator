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

	deployment := appsv1beta1.Deployment{}
	deployment.Name = applicationName
	deployment.ObjectMeta.Annotations = make(map[string]string)
	deployment.ObjectMeta.Annotations[AnnotationToWatch] = "true"

	srv := NewIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	newDeployment, _ := srv.Annotate(deployment)

	assert := assert.New(t)
	assert.Contains(newDeployment.Spec.Template.ObjectMeta.Annotations, Kube2IAMAnnotation, "The deployment PodSpec must contain the kube2iam annotation")
	assert.Equal(expectedAnnotationValue, newDeployment.Spec.Template.ObjectMeta.Annotations[Kube2IAMAnnotation], "The kube2iam annotation must contain the right IAM Role")
}

func TestDeploymentDoesntWantToBeAnnotatedThenNothingHappens(t *testing.T) {
	awsAccountID := "12345"
	applicationName := "app-name"

	deployment := appsv1beta1.Deployment{}
	deployment.Name = applicationName

	srv := NewIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	newDeployment, _ := srv.Annotate(deployment)

	assert := assert.New(t)
	assert.NotContains(newDeployment.Spec.Template.ObjectMeta.Annotations, Kube2IAMAnnotation, "The deployment PodSpec must not contain the kube2iam annotation")
}

func TestDeploymentIsSkippedBecauseAlreadyHasKube2IAMAnnotation(t *testing.T) {
	awsAccountID := "12345"
	applicationName := "app-name"

	deployment := appsv1beta1.Deployment{}
	deployment.Name = applicationName
	deployment.ObjectMeta.Annotations = make(map[string]string)
	deployment.ObjectMeta.Annotations[AnnotationToWatch] = "true"
	deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	deployment.Spec.Template.ObjectMeta.Annotations[Kube2IAMAnnotation] = "arn:aws:iam::" + awsAccountID + ":role/" + applicationName

	srv := NewIamRoleAnnotator(fake.NewSimpleClientset(), awsAccountID, NewDummyLogger())
	srv.Annotate(deployment)
}
