package zitadel

import "github.com/caos/orbos/internal/operator/boom/application/applications/logcollection/logging"

func GetFlows(outputs, clusterOutputs []string) []*logging.FlowConfig {
	return []*logging.FlowConfig{
		getOperatorFlow(outputs, clusterOutputs),
		getZitadelFlow(outputs, clusterOutputs),
	}
}

func getZitadelFlow(outputs, clusterOutputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:           "flow-zitadel",
		Namespace:      "caos-system",
		SelectLabels:   getApplicationServiceLabels(),
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}

func getOperatorFlow(outputs, clusterOutputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:           "flow-zitadel-operator",
		Namespace:      "caos-system",
		SelectLabels:   getOperatorServiceLabels(),
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}

func GetCloudZitadelFlow(outputs, clusterOutputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:             "flow-cloud-zitadel",
		Namespace:        "caos-system",
		SelectLabels:     getApplicationServiceLabels(),
		Outputs:          outputs,
		ClusterOutputs:   clusterOutputs,
		ParserType:       "regexp",
		ParserExpression: "/^.*time=(?<time>[^ ]*).*level=(?<level>[^ ]*).*logID=(?<id>[^ ]*).*$/",
		// WORKAROUND
		// According to /fluentd/log/out, only 15 labels are allowed.
		// I had to exclude the container_name as I couldn't figure
		// out how to exclude certain kubernetes labels when
		// clusteroutputs extract_kubernetes_labels is set to true.
		// Also I couldn't set it to false and explicitly include
		// certain labels.
		RemoveKeys: "$.kubernetes.docker_id,$.kubernetes.pod_id,$.kubernetes.container_name",
	}
}
