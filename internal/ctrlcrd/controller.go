package ctrlcrd

import (
	"fmt"

	boomv1 "github.com/caos/orbos/internal/api/boom/v1"
	networkingv1 "github.com/caos/orbos/internal/api/networking/v1"
	"github.com/caos/orbos/internal/ctrlcrd/boom"
	"github.com/caos/orbos/internal/ctrlcrd/networking"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	Networking = "networking"
	Boom       = "boom"
)

func Start(monitor mntr.Monitor, version, toolsDirectoryPath, metricsAddr string, kubeconfig string, features ...string) error {

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return fmt.Errorf("adding clientgo to scheme failed: %w", err)
	}

	if err := networkingv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("adding networking v1 to scheme failed: %w", err)
	}

	if err := boomv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("adding boom v1 to scheme failed: %w", err)
	}

	cfg, err := clientgo.GetClusterConfig(monitor, kubeconfig)
	if err != nil {
		return err
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     false,
		LeaderElectionID:   "98jasd12l.caos.ch",
	})
	if err != nil {
		return fmt.Errorf("starting manager failed: %w", err)
	}

	for _, feature := range features {
		switch feature {
		case Networking:
			if err = (&networking.Reconciler{
				ClientInt: kubernetes.NewK8sClientWithConfig(monitor, cfg),
				Monitor:   monitor,
				Scheme:    mgr.GetScheme(),
				Version:   version,
			}).SetupWithManager(mgr); err != nil {
				return fmt.Errorf("creating controller failed: %w", err)
			}
		case Boom:
			if err = (&boom.Reconciler{
				ClientInt:          kubernetes.NewK8sClientWithConfig(monitor, cfg),
				Monitor:            monitor,
				Scheme:             mgr.GetScheme(),
				ToolsDirectoryPath: toolsDirectoryPath,
				Version:            version,
			}).SetupWithManager(mgr); err != nil {
				return fmt.Errorf("creating controller failed: %w", err)
			}
		}
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("running manager failed: %w", err)
	}
	return nil
}
