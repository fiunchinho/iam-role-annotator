package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spotahome/kooper/log"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"time"
	"github.com/fiunchinho/iam-role-annotator/service"
	"github.com/fiunchinho/iam-role-annotator/controller"
)

func getKubernetesClient(kubeconfig string) (kubernetes.Interface, error) {
	var err error
	var cfg *rest.Config

	cfg, err = rest.InClusterConfig()
	if err != nil {
		fmt.Printf("Falling back to using kubeconfig file: %s\n", err)
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("could not load configuration: %s", err)
		}
	}

	return kubernetes.NewForConfig(cfg)
}

func main() {
	logger := &log.Std{}

	stopC := make(chan struct{})
	finishC := make(chan error)
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGTERM, syscall.SIGINT)

	// Parse command line arguments.
	flags, err := NewFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %s", err)
		os.Exit(1)
	}

	// Get kubernetes rest client.
	k8sCli, err := getKubernetesClient(flags.KubeConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't create k8s client: %s", err)
		os.Exit(1)
	}

	// Create the controller and run.
	ctrl, err := controller.New(
		time.Duration(flags.ResyncSec)*time.Second,
		controller.NewHandler(*service.NewIamRoleAnnotator(k8sCli, flags.AWSAccountID, logger)),
		controller.NewDeploymentRetrieve(flags.Namespace, k8sCli),
		logger,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}

	// Run in background the controller.
	go func() {
		finishC <- ctrl.Run(stopC)
	}()

	select {
	case err := <-finishC:
		if err != nil {
			fmt.Fprintf(os.Stderr, "error running controller: %s", err)
			os.Exit(1)
		}
	case <-signalC:
		logger.Infof("Signal captured, exiting...")
	}

}
