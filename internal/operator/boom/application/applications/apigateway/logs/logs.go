package logs

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/logcollection/logging"
)

func getLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/instance": "ambassador",
		"app.kubernetes.io/name":     "ambassador",
	}
}

func GetFlow(outputs []string, clusterOutputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:             "flow-ambassador",
		Namespace:        "caos-system",
		SelectLabels:     getLabels(),
		Outputs:          outputs,
		ClusterOutputs:   clusterOutputs,
		ParserType:       "logfmt",
		ParserExpression: "",
	}
}

func GetCloudFlow(outputs []string, clusterOutputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:             "flow-cloud-ambassador",
		Namespace:        "caos-system",
		SelectLabels:     getLabels(),
		Outputs:          outputs,
		ClusterOutputs:   clusterOutputs,
		ParserType:       "regexp",
		ParserExpression: "^(?<type>\\S+) \\[(?<time>[^\\]]*)\\] \"(?<method>\\S+)(?: +(?<path>(?:[^\\\"]|\\\\.)*?)(?: +\\S*)?) (?<protocol>\\S+)?\" (?<response_code>\\S+) (?<response_flags>\\S+) (?<bytes_received>\\S+) (?<bytes_sent>\\S+) (?<duration>\\S+) (?<x_envoy_upstream_service_time>\\S+) \"(?<x_forwarded_for>[^\\\"]*)\" \"(?<user_agent>[^\\\"]*)\" \"(?<x_request_id>[^\\\"]*)\" \"(?<authority>[^\\\"]*)\" \"(?<upstream_host>[^\\\"]*)\"",
	}
}
