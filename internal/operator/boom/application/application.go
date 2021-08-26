package application

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/apigateway"
	apigatewayinfo "github.com/caos/orbos/internal/operator/boom/application/applications/apigateway/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/kubemetricsexporter"
	kubemetricsexporterinfo "github.com/caos/orbos/internal/operator/boom/application/applications/kubemetricsexporter/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/logcollection"
	logcollectioninfo "github.com/caos/orbos/internal/operator/boom/application/applications/logcollection/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/logspersisting"
	logspersistinginfo "github.com/caos/orbos/internal/operator/boom/application/applications/logspersisting/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metriccollection"
	metriccollectioninfo "github.com/caos/orbos/internal/operator/boom/application/applications/metriccollection/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting"
	metricspersistinginfo "github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricsserver"
	metricsserverinfo "github.com/caos/orbos/internal/operator/boom/application/applications/metricsserver/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/monitoring"
	monitoringinfo "github.com/caos/orbos/internal/operator/boom/application/applications/monitoring/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/nodemetricsexporter"
	nodemetricsexporterinfo "github.com/caos/orbos/internal/operator/boom/application/applications/nodemetricsexporter/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/reconciling"
	reconcilinginfo "github.com/caos/orbos/internal/operator/boom/application/applications/reconciling/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/systemdmetricsexporter"
	systemdmetricsexporterinfo "github.com/caos/orbos/internal/operator/boom/application/applications/systemdmetricsexporter/info"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
)

type Application interface {
	Deploy(*latest.ToolsetSpec) bool
	GetName() name.Application
}

type HelmApplication interface {
	Application
	GetNamespace() string
	GetChartInfo() *chart.Chart
	GetImageTags() map[string]string
	SpecToHelmValues(mntr.Monitor, *latest.ToolsetSpec) interface{}
}

type YAMLApplication interface {
	Application
	GetYaml(mntr.Monitor, *latest.ToolsetSpec) interface{}
}

func New(monitor mntr.Monitor, appName name.Application, orb string) Application {
	switch appName {
	case apigatewayinfo.GetName():
		return apigateway.New(monitor)
	case reconcilinginfo.GetName():
		return reconciling.New(monitor)
	case monitoringinfo.GetName():
		return monitoring.New(monitor)
	case kubemetricsexporterinfo.GetName():
		return kubemetricsexporter.New(monitor)
	case metriccollectioninfo.GetName():
		return metriccollection.New(monitor)
	case logcollectioninfo.GetName():
		return logcollection.New(monitor, orb)
	case nodemetricsexporterinfo.GetName():
		return nodemetricsexporter.New(monitor)
	case systemdmetricsexporterinfo.GetName():
		return systemdmetricsexporter.New()
	case metricspersistinginfo.GetName():
		return metricspersisting.New(monitor, orb)
	case logspersistinginfo.GetName():
		return logspersisting.New(monitor)
	case metricsserverinfo.GetName():
		return metricsserver.New(monitor)
	}

	return nil
}

func GetOrderNumber(appName name.Application) int {
	switch appName {
	case metricspersistinginfo.GetName():
		return metricspersistinginfo.GetOrderNumber()
	case logspersistinginfo.GetName():
		return logspersistinginfo.GetOrderNumber()
	}

	return 1
}
