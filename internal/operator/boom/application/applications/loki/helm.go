package loki

import (
	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loki/helm"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loki/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loki/logs"
	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
)

func (l *Loki) HelmPreApplySteps(monitor mntr.Monitor, toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec) ([]interface{}, error) {
	return logs.GetAllResources(toolsetCRDSpec), nil
}

func (l *Loki) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetsv1beta1.ToolsetSpec) interface{} {
	spec := toolset.Loki
	values := helm.DefaultValues(l.GetImageTags())

	if spec.Storage != nil {
		values.Persistence.Enabled = true
		values.Persistence.Size = spec.Storage.Size
		values.Persistence.StorageClassName = spec.Storage.StorageClass
		if spec.Storage.AccessModes != nil {
			values.Persistence.AccessModes = spec.Storage.AccessModes
		}
	}

	values.FullNameOverride = info.GetName().String()
	return values
}

func (l *Loki) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (l *Loki) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
