package networking

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	operatorLabels *labels.Operator,
	desiredTree *tree.Tree,
	currentTree *tree.Tree,
	namespace string,
) (
	query core.QueryFunc,
	destroy core.DestroyFunc,
	secrets map[string]*secret.Secret,
	existing map[string]*secret.Existing,
	err error,
) {
	switch desiredTree.Common.Kind {
	case "networking.caos.ch/LegacyCloudflare":
		return legacycf.AdaptFunc(namespace, operatorLabels)(monitor, desiredTree, currentTree)
	default:
		return nil, nil, nil, nil, errors.Errorf("unknown networking kind %s", desiredTree.Common.Kind)
	}
}
