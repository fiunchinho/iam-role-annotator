package main

import (
	"os"
	"os/signal"
	"syscall"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"time"

	"github.com/fiunchinho/iam-role-annotator/internal"
	"github.com/fiunchinho/iam-role-annotator/pkg"
	"github.com/spotahome/kooper/operator/controller"
	"go.uber.org/zap"
)

func getKubernetesClient(kubeconfig string, logger pkg.Logger) (kubernetes.Interface, error) {
	var err error
	var cfg *rest.Config

	cfg, err = rest.InClusterConfig()
	if err != nil {
		logger.Warningf("Falling back to using kubeconfig file: %s", err)
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	return kubernetes.NewForConfig(cfg)
}

func main() {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	logger := pkg.NewLogger(*zapLogger.Sugar())

	stopC := make(chan struct{})
	finishC := make(chan error)
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGTERM, syscall.SIGINT)

	// Parse command line arguments.
	flags, err := NewFlags()
	if err != nil {
		logger.Errorf("Invalid configuration: %s", err)
		os.Exit(1)
	}

	// Get kubernetes rest client.
	k8sCli, err := getKubernetesClient(flags.KubeConfig, logger)
	if err != nil {
		logger.Errorf("Can't create k8s client: %s", err)
		os.Exit(1)
	}

	// Create the controller and run.
	ctrl := controller.NewSequential(
		time.Duration(flags.ResyncSec)*time.Second,
		internal.NewHandler(*pkg.NewIamRoleAnnotator(k8sCli, flags.AWSAccountID, logger)),
		internal.NewDeploymentRetrieve(flags.Namespace, k8sCli),
		nil,
		logger,
	)

	// Run in background the controller.
	go func() {
		finishC <- ctrl.Run(stopC)
	}()

	select {
	case err := <-finishC:
		if err != nil {
			logger.Errorf("error running controller: %s", err)
			os.Exit(1)
		}
	case <-signalC:
		logger.Info("Signal captured, exiting...")
	}
}
