package networking

import (
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/legacycf"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	desiredTree *tree.Tree,
	currentTree *tree.Tree,
	namespace string,
	labels map[string]string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	map[string]*secret.Secret,
	error,
) {
	switch desiredTree.Common.Kind {
	case "zitadel.caos.ch/LegacyCloudflare":
		return legacycf.AdaptFunc(namespace, labels)(monitor, desiredTree, currentTree)
	default:
		return nil, nil, nil, errors.Errorf("unknown networking kind %s", desiredTree.Common.Kind)
	}
}
