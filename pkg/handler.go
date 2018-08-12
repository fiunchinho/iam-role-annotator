package pkg

import (
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Handler will receive objects whenever they get added or deleted from the k8s API.
type Handler struct {
	iamRoleAnnotator IamRoleAnnotatorInterface
}

// Add is called when a k8s object is created.
func (h *Handler) Add(obj runtime.Object) error {
	_, err := h.iamRoleAnnotator.Annotate(*obj.(*appsv1beta1.Deployment))
	return err
}

// Delete is called when a k8s object is deleted.
func (h *Handler) Delete(s string) error {
	return nil
}

// NewHandler returns a new Handler to handle Deployments created/updated/deleted.
func NewHandler(iamRoleAnnotator IamRoleAnnotatorInterface) *Handler {
	return &Handler{
		iamRoleAnnotator: iamRoleAnnotator,
	}
}
