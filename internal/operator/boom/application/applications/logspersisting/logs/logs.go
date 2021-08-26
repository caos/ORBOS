package logs

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/boom"
	"github.com/caos/orbos/internal/operator/boom/application/applications/database"
	"github.com/caos/orbos/internal/operator/boom/application/applications/networking"
	"github.com/caos/orbos/internal/operator/boom/application/applications/orbiter"
	"github.com/caos/orbos/internal/operator/boom/application/applications/zitadel"
	"strings"

	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	amlogs "github.com/caos/orbos/internal/operator/boom/application/applications/apigateway/logs"

	ksmlogs "github.com/caos/orbos/internal/operator/boom/application/applications/kubemetricsexporter/logs"
	"github.com/caos/orbos/internal/operator/boom/application/applications/logcollection/logging"
	lologs "github.com/caos/orbos/internal/operator/boom/application/applications/logcollection/logs"
	"github.com/caos/orbos/internal/operator/boom/application/applications/logspersisting/info"
	pologs "github.com/caos/orbos/internal/operator/boom/application/applications/metriccollection/logs"
	plogs "github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/logs"
	mslogs "github.com/caos/orbos/internal/operator/boom/application/applications/metricsserver/logs"
	glogs "github.com/caos/orbos/internal/operator/boom/application/applications/monitoring/logs"
	pnelogs "github.com/caos/orbos/internal/operator/boom/application/applications/nodemetricsexporter/logs"
	aglogs "github.com/caos/orbos/internal/operator/boom/application/applications/reconciling/logs"
	pselogs "github.com/caos/orbos/internal/operator/boom/application/applications/systemdmetricsexporter/logs"
	"github.com/caos/orbos/internal/operator/boom/labels"
)

func GetAllResources(
	toolsetCRDSpec *toolsetslatest.ToolsetSpec,
	withGrafanaCloud bool,
	secretName string,
	orb string,
) []interface{} {

	if toolsetCRDSpec.LogCollection == nil || !toolsetCRDSpec.LogCollection.Deploy {
		return nil
	}

	ret := []interface{}{logging.New(toolsetCRDSpec.LogCollection)}

	outputNames := toolsetCRDSpec.LogCollection.Outputs
	clusterOutputNames := toolsetCRDSpec.LogCollection.ClusterOutputs
	var outputs []*logging.Output
	// output to loki
	if toolsetCRDSpec.LogsPersisting != nil && toolsetCRDSpec.LogsPersisting.Deploy {
		lokiOutputNames, lokiClusterOutputNames, lokiOutputs := getLokiOutput(toolsetCRDSpec.LogsPersisting.ClusterOutput)
		outputNames = append(outputNames, lokiOutputNames...)
		clusterOutputNames = append(clusterOutputNames, lokiClusterOutputNames...)
		outputs = append(outputs, lokiOutputs...)
	}

	if withGrafanaCloud {
		lokiOutputNames, lokiClusterOutputNames, lokiOutputs := getCloudLokiOutput(toolsetCRDSpec.LogsPersisting.ClusterOutput, secretName, orb)
		outputs = append(outputs, lokiOutputs...)

		if len(lokiOutputNames) > 0 || len(lokiClusterOutputNames) > 0 {
			flows := getAllCloudFlows(toolsetCRDSpec, lokiOutputNames, lokiClusterOutputNames)
			for _, flow := range flows {
				ret = append(ret, flow)
			}
		}
	}

	for _, output := range outputs {
		ret = append(ret, output)
	}

	// add flows for each application
	if len(outputNames) > 0 || len(clusterOutputNames) > 0 {
		flows := getAllFlows(toolsetCRDSpec, outputNames, clusterOutputNames)
		for _, flow := range flows {
			ret = append(ret, flow)
		}
	}

	return ret
}

func getAllCloudFlows(
	toolsetCRDSpec *toolsetslatest.ToolsetSpec,
	cloudOutputNames []string,
	cloudClusterOutputs []string,
) []*logging.Flow {

	if toolsetCRDSpec.LogsPersisting == nil || !toolsetCRDSpec.LogsPersisting.Deploy {
		return []*logging.Flow{}
	}

	flows := make([]*logging.Flow, 0)
	if toolsetCRDSpec.APIGateway != nil && toolsetCRDSpec.APIGateway.Deploy {
		flows = append(flows, logging.NewFlow(amlogs.GetCloudFlow(cloudOutputNames, cloudClusterOutputs)))
	}

	if toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Zitadel {
		flows = append(flows, logging.NewFlow(zitadel.GetCloudZitadelFlow(cloudOutputNames, cloudClusterOutputs)))
	}

	return flows
}

func getAllFlows(
	toolsetCRDSpec *toolsetslatest.ToolsetSpec,
	outputNames []string,
	clusterOutputs []string,
) []*logging.Flow {

	if toolsetCRDSpec.LogsPersisting == nil || !toolsetCRDSpec.LogsPersisting.Deploy {
		return []*logging.Flow{}
	}

	flows := make([]*logging.Flow, 0)
	if toolsetCRDSpec.APIGateway != nil && toolsetCRDSpec.APIGateway.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Ambassador) {
		flows = append(flows, logging.NewFlow(amlogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.Monitoring != nil && toolsetCRDSpec.Monitoring.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Grafana) {
		flows = append(flows, logging.NewFlow(glogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.MetricCollection != nil && toolsetCRDSpec.MetricCollection.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.PrometheusOperator) {
		flows = append(flows, logging.NewFlow(pologs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.NodeMetricsExporter != nil && toolsetCRDSpec.NodeMetricsExporter.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.PrometheusNodeExporter) {
		flows = append(flows, logging.NewFlow(pnelogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.KubeMetricsExporter != nil && toolsetCRDSpec.KubeMetricsExporter.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.KubeStateMetrics) {
		flows = append(flows, logging.NewFlow(ksmlogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.Reconciling != nil && toolsetCRDSpec.Reconciling.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Argocd) {
		flows = append(flows, logging.NewFlow(aglogs.GetFlow(outputNames, clusterOutputs)))
	}
	if toolsetCRDSpec.LogCollection != nil && toolsetCRDSpec.LogCollection.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.LoggingOperator) {
		flows = append(flows, logging.NewFlow(lologs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.MetricsPersisting != nil && toolsetCRDSpec.MetricsPersisting.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Prometheus) {
		flows = append(flows, logging.NewFlow(plogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.MetricsServer != nil && toolsetCRDSpec.MetricsServer.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.MetricsServer) {
		flows = append(flows, logging.NewFlow(mslogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.SystemdMetricsExporter != nil && toolsetCRDSpec.SystemdMetricsExporter.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.PrometheusSystemdExporter) {
		flows = append(flows, logging.NewFlow(pselogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Loki {
		flows = append(flows, logging.NewFlow(getLokiFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Boom {
		flows = append(flows, logging.NewFlow(boom.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Orbiter {
		flows = append(flows, logging.NewFlow(orbiter.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Zitadel {
		for _, flow := range zitadel.GetFlows(outputNames, clusterOutputs) {
			flows = append(flows, logging.NewFlow(flow))
		}
	}

	if toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Database {
		for _, flow := range database.GetFlows(outputNames, clusterOutputs) {
			flows = append(flows, logging.NewFlow(flow))
		}
	}

	if toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Networking {
		flows = append(flows, logging.NewFlow(networking.GetFlow(outputNames, clusterOutputs)))
	}

	return flows
}

func getLokiFlow(outputs []string, clusterOutputs []string) *logging.FlowConfig {
	ls := labels.GetApplicationLabels(info.GetName())

	return &logging.FlowConfig{
		Name:           "flow-loki",
		Namespace:      "caos-system",
		SelectLabels:   ls,
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}

func getLokiOutput(clusterOutput bool) ([]string, []string, []*logging.Output) {
	outputURL := strings.Join([]string{"http://", info.GetName().String(), ".", info.GetNamespace(), ":3100"}, "")

	conf := &logging.ConfigOutput{
		Name:      "output-loki",
		Namespace: "caos-system",
		URL:       outputURL,
	}

	outputs := make([]*logging.Output, 0)
	outputs = append(outputs, logging.NewOutput(clusterOutput, conf))

	outputNames := make([]string, 0)
	clusterOutputs := make([]string, 0)
	if clusterOutput {
		clusterOutputs = append(clusterOutputs, conf.Name)
	} else {
		outputNames = append(outputNames, conf.Name)
	}

	return outputNames, clusterOutputs, outputs
}

func getCloudLokiOutput(
	clusterOutput bool,
	secretName string,
	orb string,
) ([]string, []string, []*logging.Output) {
	conf := &logging.ConfigOutput{
		EnabledNamespaces: []string{
			"caos-system",
			"caos-zitadel",
		},
		ConfigureKubernetesLabels: true,
		ExtractKubernetesLabels:   true,
		ExtraLabels: map[string]string{
			"orb": orb,
		},
		Labels: map[string]string{
			"id":    "$.id",
			"level": "$.level",
		},
		RemoveKeys: []string{
			"kubernetes",
		},
		Name:      "output-grafana-cloud-loki",
		Namespace: "caos-system",
		URL:       "https://logs-prod-us-central1.grafana.net",
		Username: &logging.SecretKeyRef{
			Key:  "username",
			Name: secretName,
		},
		Password: &logging.SecretKeyRef{
			Key:  "password",
			Name: secretName,
		},
	}

	outputs := make([]*logging.Output, 0)
	outputs = append(outputs, logging.NewOutput(clusterOutput, conf))

	outputNames := make([]string, 0)
	clusterOutputs := make([]string, 0)
	if clusterOutput {
		clusterOutputs = append(clusterOutputs, conf.Name)
	} else {
		outputNames = append(outputNames, conf.Name)
	}

	return outputNames, clusterOutputs, outputs
}
