package pkg

import (
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// DeploymentRetrieve knows how to retrieve Deployments.
type DeploymentRetrieve struct {
	namespace string
	client    kubernetes.Interface
}

// NewDeploymentRetrieve returns a new Deployment retriever.
func NewDeploymentRetrieve(namespace string, client kubernetes.Interface) *DeploymentRetrieve {
	return &DeploymentRetrieve{
		namespace: namespace,
		client:    client,
	}
}

// GetListerWatcher knows how to return a listerWatcher of a Deployment.
func (p *DeploymentRetrieve) GetListerWatcher() cache.ListerWatcher {

	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return p.client.AppsV1beta1().Deployments(p.namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return p.client.AppsV1beta1().Deployments(p.namespace).Watch(options)
		},
	}
}

// GetObject returns an empty Deployment.
func (p *DeploymentRetrieve) GetObject() runtime.Object {
	return &appsv1beta1.Deployment{}
}
