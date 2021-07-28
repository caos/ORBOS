package main

import (
	"context"
	"flag"
	"os"

	"github.com/caos/orbos/internal/ctrlcrd"
	"github.com/caos/orbos/internal/ctrlgitops"
	"github.com/caos/orbos/pkg/git"
	"github.com/caos/orbos/pkg/kubernetes/cli"

	"github.com/caos/orbos/pkg/orb"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/mntr"
)

func main() {
	var orbconfig, kubeconfig, metricsAddr string
	var verbose, gitopsmode bool

	flag.StringVar(&orbconfig, "orbconfig", "~/.orb/config", "The orbconfig file to use")
	flag.StringVar(&kubeconfig, "kc", "~/.kube/config", "The kubeconfig file to use")
	flag.BoolVar(&verbose, "verbose", false, "Print debug levelled logs")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&gitopsmode, "gitopsmode", false, "defines if the operator should run in gitops mode not crd mode")

	flag.Parse()

	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	if verbose {
		monitor = monitor.Verbose()
	}

	prunedPath := helpers.PruneHome(orbconfig)
	orbConfig, err := orb.ParseOrbConfig(prunedPath)
	if err != nil {
		monitor.Error(err)
		os.Exit(1)
	}
	gitClient := git.New(context.Background(), monitor, "orbos", "orbos@caos.ch")
	kubeconfig = helpers.PruneHome(kubeconfig)
	version := "networking-development"

	if gitopsmode {

		k8sClient, err := cli.Init(monitor, orbConfig, gitClient, kubeconfig, gitopsmode, true, false)
		if err != nil {
			monitor.Error(err)
			os.Exit(1)
		}
		if err := ctrlgitops.Networking(monitor, orbConfig.Path, k8sClient, &version); err != nil {
			monitor.Error(err)
			os.Exit(1)
		}
	} else {
		if err := ctrlcrd.Start(monitor, version, "/boom", metricsAddr, kubeconfig, ctrlcrd.Networking); err != nil {
			monitor.Error(err)
			os.Exit(1)
		}
	}
}
