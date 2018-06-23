package controller

import (
	"time"

	"github.com/spotahome/kooper/log"
	"github.com/spotahome/kooper/operator/controller"
)

// Controller is a controller that annotates Deployments.
type Controller struct {
	controller.Controller
	logger log.Logger
}

// New returns a new Iam Role Annotator controller.
func New(resyncPeriod time.Duration, handler *Handler, retriever *DeploymentRetrieve, logger log.Logger) (*Controller, error) {
	kooperController := controller.NewSequential(
		resyncPeriod,
		handler,
		retriever,
		nil,
		logger)

	return &Controller{
		Controller: kooperController,
		logger:     logger,
	}, nil
}
