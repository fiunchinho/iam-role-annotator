package cmd

import (
	"fmt"
	"os"

	"path/filepath"

	"os/signal"
	"syscall"

	"time"

	"k8s.io/client-go/util/homedir"

	"github.com/fiunchinho/iam-role-annotator/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/spotahome/kooper/operator/controller"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	zapLogger, _ = zap.NewProduction(zap.AddCallerSkip(1))
	logger       = pkg.NewLogger(*zapLogger.Sugar())
	stopC        = make(chan struct{})
	finishC      = make(chan error)
	signalC      = make(chan os.Signal, 1)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iam-role-annotator",
	Short: "Kubernetes controller that automatically adds annotations in Pods to they can assume AWS Roles.",
	Long:  `Kubernetes controller that automatically adds annotations in Pods to they can assume AWS Roles.`,
	Run: func(cmd *cobra.Command, args []string) {
		defer zapLogger.Sync()
		signal.Notify(signalC, syscall.SIGTERM, syscall.SIGINT)

		k8sCli, err := getKubernetesClient(viper.GetString("kubeconfig"), logger)
		if err != nil {
			logger.Errorf("Can't create k8s client: %s", err)
			os.Exit(1)
		}

		go func() {
			finishC <- getController(k8sCli, logger).Run(stopC)
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
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		namespace := viper.GetString("namespace")
		if len(namespace) == 0 {
			logger.Error("Error: required flag \"namespace\" or environment variable \"NAMESPACE\" not set")
			os.Exit(1)
		}

		awsAccountID := viper.GetString("aws-account-id")
		if len(awsAccountID) == 0 {
			logger.Error("Error: required flag \"aws-account-id\" or environment variable \"AWS_ACCOUNT_ID\" not set")
			os.Exit(1)
		}
	})

	rootCmd.Flags().String("kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "Path to the kubeconfig file")
	viper.BindPFlag("kubeconfig", rootCmd.Flags().Lookup("kubeconfig"))

	rootCmd.Flags().String("namespace", "", "kubernetes namespace where this app is running")
	viper.BindPFlag("namespace", rootCmd.Flags().Lookup("namespace"))

	rootCmd.Flags().String("aws-account-id", "", "AWS Account ID that appears on the ARN of the roles referenced on the annotations")
	viper.BindPFlag("aws-account-id", rootCmd.Flags().Lookup("aws-account-id"))
	viper.BindEnv("aws-account-id", "AWS_ACCOUNT_ID")

	rootCmd.Flags().Int("resync-seconds", 30, "The number of seconds the controller will resync the resources")
	viper.BindPFlag("resync-seconds", rootCmd.Flags().Lookup("resync-seconds"))

	rootCmd.Flags().Bool("enable-opt-out-mode", false, "Enable the opt-out mode, where all Deployments will get the IAM annotation unless they explicitly opt out of it")
	viper.BindPFlag("enable-opt-out-mode", rootCmd.Flags().Lookup("enable-opt-out-mode"))

	viper.AutomaticEnv()
}

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

func getController(k8sCli kubernetes.Interface, logger pkg.Logger) controller.Controller {
	annotator := pkg.NewOptinIamRoleAnnotator(k8sCli, viper.GetString("aws-account-id"), logger)
	if viper.GetBool("enable-opt-out-mode") {
		annotator = pkg.NewOptoutIamRoleAnnotator(k8sCli, viper.GetString("aws-account-id"), logger)
	}

	return controller.NewSequential(
		time.Duration(viper.GetInt("resync-seconds"))*time.Second,
		pkg.NewHandler(annotator),
		pkg.NewDeploymentRetrieve(viper.GetString("namespace"), k8sCli),
		nil,
		logger,
	)
}
