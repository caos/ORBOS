package orbiter

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/pkg/labels"
)

func GetFlow(outputs, clusterOutputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:           "flow-orbiter",
		Namespace:      "caos-system",
		SelectLabels:   labels.MustK8sMap(orb.OperatorSelector()),
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}
