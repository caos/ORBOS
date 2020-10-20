package legacycf

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/core"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
) zitadel.AdaptFunc {
	return func(
		monitor mntr.Monitor,
		desiredTree *tree.Tree,
		currentTree *tree.Tree,
	) (
		zitadel.QueryFunc,
		zitadel.DestroyFunc,
		map[string]*secret.Secret,
		error,
	) {
		internalMonitor := monitor.WithField("kind", "legacycf")

		desiredKind, err := parseDesired(desiredTree)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		if !monitor.IsVerbose() && desiredKind.Spec.Verbose {
			internalMonitor.Verbose()
		}

		if desiredKind.Spec == nil {
			return nil, nil, nil, errors.New("No specs found")
		}

		if err := desiredKind.Spec.Validate(); err != nil {
			return nil, nil, nil, err
		}

		internalSpec, current := desiredKind.Spec.Internal(namespace, labels)

		legacyQuerier, legacyDestroyer, readyCertificate, err := adaptFunc(monitor, internalSpec)
		current.ReadyCertificate = readyCertificate

		queriers := []zitadel.QueryFunc{
			legacyQuerier,
		}
		currentTree.Parsed = current

		return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
				core.SetQueriedForNetworking(queried, currentTree)
				internalMonitor.Info("set current state legacycf")

				return zitadel.QueriersToEnsureFunc(internalMonitor, true, queriers, k8sClient, queried)
			},
			zitadel.DestroyersToDestroyFunc(internalMonitor, []zitadel.DestroyFunc{legacyDestroyer}),
			getSecretsMap(desiredKind),
			nil
	}
}
