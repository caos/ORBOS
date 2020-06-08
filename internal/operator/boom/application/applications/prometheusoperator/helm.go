package prometheusoperator

import (
	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheusoperator/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
)

func (p *PrometheusOperator) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetsv1beta1.ToolsetSpec) interface{} {
	// spec := toolset.PrometheusNodeExporter
	values := helm.DefaultValues(p.GetImageTags())

	// if spec.ReplicaCount != 0 {
	// 	values.ReplicaCount = spec.ReplicaCount
	// }

	return values
}

func (p *PrometheusOperator) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (p *PrometheusOperator) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
