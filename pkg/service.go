package pkg

import (
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/client-go/kubernetes"
)

const (
	// AnnotationToWatch is the annotation that Deployments need to have
	AnnotationToWatch = "armesto.net/iam-role-annotator"
	// Kube2IAMAnnotation is the standard kube2iam annotation used to specify the IAM Role to assume
	Kube2IAMAnnotation = "iam.amazonaws.com/role"
)

// IamRoleAnnotatorInterface for the IamRoleAnnotator
type IamRoleAnnotatorInterface interface {
	Annotate(deployment appsv1beta1.Deployment) (*appsv1beta1.Deployment, error)
}

// IamRoleAnnotator is simple annotator service.
type IamRoleAnnotator struct {
	client       kubernetes.Interface
	awsAccountID string
	logger       Logger
}

// NewIamRoleAnnotator returns a new IamRoleAnnotator.
func NewIamRoleAnnotator(k8sCli kubernetes.Interface, awsAccountID string, logger Logger) *IamRoleAnnotator {
	return &IamRoleAnnotator{
		client:       k8sCli,
		awsAccountID: awsAccountID,
		logger:       logger,
	}
}

// Annotate will add the kube2iam annotation to Deployment objects containing the special annotation
func (s *IamRoleAnnotator) Annotate(deployment appsv1beta1.Deployment) (*appsv1beta1.Deployment, error) {
	newDeployment := deployment.DeepCopy()

	_, deploymentExpectsToBeAnnotated := newDeployment.ObjectMeta.Annotations[AnnotationToWatch]
	if !deploymentExpectsToBeAnnotated {
		return newDeployment, nil
	}

	s.logger.Infof("Detected deploy/%s annotated with %s: %s", newDeployment.ObjectMeta.Name, AnnotationToWatch, newDeployment.ObjectMeta.Annotations[AnnotationToWatch])

	_, alreadyContainsKube2IamAnnotation := newDeployment.Spec.Template.ObjectMeta.Annotations[Kube2IAMAnnotation]
	if alreadyContainsKube2IamAnnotation {
		s.logger.Infof("deploy/%s already contains %s:%s, skipping...", newDeployment.ObjectMeta.Name, Kube2IAMAnnotation, newDeployment.Spec.Template.ObjectMeta.Annotations[Kube2IAMAnnotation])
		return newDeployment, nil
	}

	s.annotate(newDeployment, s.getRoleArn(newDeployment))
	_, err := s.submitChangesToKubernetesAPI(newDeployment)

	return newDeployment, err
}

// getRoleArn returns the role arn to put in the Deployment annotation.
func (s *IamRoleAnnotator) getRoleArn(deployment *appsv1beta1.Deployment) string {
	return "arn:aws:iam::" + s.awsAccountID + ":role/" + deployment.ObjectMeta.Name
}

// annotate adds the annotation to the Deployment
func (s *IamRoleAnnotator) annotate(deployment *appsv1beta1.Deployment, role string) {
	s.logger.Infof("Adding IAM Role annotation '%s' to the PodSpec of the Deployment '%s'", role, deployment.Name)
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.ObjectMeta.Annotations[Kube2IAMAnnotation] = role
}

// submitChangesToKubernetesAPI updates the Deployment object in the Kubernetes API
func (s *IamRoleAnnotator) submitChangesToKubernetesAPI(deployment *appsv1beta1.Deployment) (*appsv1beta1.Deployment, error) {
	s.logger.Infof("Sending changes to k8s API")
	return s.client.AppsV1beta1().Deployments(deployment.Namespace).Update(deployment)
}
