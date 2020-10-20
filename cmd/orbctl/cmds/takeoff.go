package cmds

import (
	"context"
	"errors"
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/start"
	"github.com/caos/orbos/mntr"
	"io/ioutil"
)

func Takeoff(
	monitor mntr.Monitor,
	ctx context.Context,
	orbConfig *orb.Orb,
	gitClient *git.Client,
	recur bool,
	destroy bool,
	deploy bool,
	verbose bool,
	ingestionAddress string,
	version string,
	gitCommit string,
	kubeconfig string,
) error {
	if err := orbConfig.IsComplete(); err != nil {
		return err
	}

	if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
		return err
	}

	if err := gitClient.Clone(); err != nil {
		return err
	}

	allKubeconfigs := make([]string, 0)
	foundOrbiter, err := api.ExistsOrbiterYml(gitClient)
	if err != nil {
		return err
	}
	if foundOrbiter {
		orbiterConfig := &start.OrbiterConfig{
			Recur:            recur,
			Destroy:          destroy,
			Deploy:           deploy,
			Verbose:          verbose,
			Version:          version,
			OrbConfigPath:    orbConfig.Path,
			GitCommit:        gitCommit,
			IngestionAddress: ingestionAddress,
		}

		kubeconfigs, err := start.Orbiter(ctx, monitor, orbiterConfig, gitClient, orbConfig)
		if err != nil {
			return err
		}
		allKubeconfigs = append(allKubeconfigs, kubeconfigs...)
	} else {
		if kubeconfig == "" {
			return errors.New("Error to deploy BOOM as no kubeconfig is provided")
		}
		value, err := ioutil.ReadFile(kubeconfig)
		if err != nil {
			return err
		}
		allKubeconfigs = append(allKubeconfigs, string(value))
	}

	for _, kubeconfig := range allKubeconfigs {
		k8sClient := kubernetes.NewK8sClient(monitor, &kubeconfig)
		if k8sClient.Available() {
			if err := kubernetes.EnsureCommonArtifacts(monitor, k8sClient); err != nil {
				monitor.Info("failed to apply common resources into k8s-cluster")
				return err
			}
			monitor.Info("Applied common resources")

			if err := kubernetes.EnsureConfigArtifacts(monitor, k8sClient, orbConfig); err != nil {
				monitor.Info("failed to apply configuration resources into k8s-cluster")
				return err
			}
			monitor.Info("Applied configuration resources")
		} else {
			monitor.Info("Failed to connect to k8s")
		}

		if err := deployBoom(monitor, gitClient, &kubeconfig, version); err != nil {
			return err
		}

		if err := deployZitadel(monitor, gitClient, &kubeconfig, version); err != nil {
			return err
		}
	}
	return nil
}
