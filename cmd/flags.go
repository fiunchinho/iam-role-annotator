package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
)

// Flags are the controller flags.
type Flags struct {
	flagSet *flag.FlagSet

	Namespace    string
	ResyncSec    int
	KubeConfig   string
	AWSAccountID string
}

// NewFlags returns a new Flags.
func NewFlags() (*Flags, error) {
	f := &Flags{
		flagSet: flag.NewFlagSet(os.Args[0], flag.ExitOnError),
	}
	// Get the user kubernetes configuration in it's home directory.
	kubehome := filepath.Join(homedir.HomeDir(), ".kube", "config")

	// Init flags.
	f.flagSet.StringVar(&f.Namespace, "namespace", "", "kubernetes namespace where this app is running")
	f.flagSet.IntVar(&f.ResyncSec, "resync-seconds", 30, "The number of seconds the controller will resync the resources")
	f.flagSet.StringVar(&f.KubeConfig, "kubeconfig", kubehome, "kubernetes configuration path, only used when running outside a kubernetes cluster")
	f.flagSet.StringVar(&f.AWSAccountID, "aws-account-id", "", "AWS Account ID that appears on the ARN of the roles referenced on the annotations")

	f.flagSet.Parse(os.Args[1:])
	flag.CommandLine.Parse([]string{})

	if len(f.Namespace) < 1 {
		return nil, fmt.Errorf("you need to pass the '--namespace' parameter")
	}
	if len(f.AWSAccountID) < 1 {
		return nil, fmt.Errorf("you need to pass the '--aws-account-id' parameter")
	}

	return f, nil
}
