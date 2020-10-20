package monitoring

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/k8s"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/admin"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/auth"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/network"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/storage"
)

type Monitoring struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Spec for the definition of the admin account
	Admin *admin.Admin `json:"admin,omitempty" yaml:"admin,omitempty"`
	//Spec for additional datasources
	Datasources []*Datasource `json:"datasources,omitempty" yaml:"datasources,omitempty"`
	//Spec for additional Dashboardproviders
	DashboardProviders []*Provider `json:"dashboardproviders,omitempty" yaml:"dashboardproviders,omitempty"`
	//Spec to define how the persistence should be handled
	Storage *storage.Spec `json:"storage,omitempty" yaml:"storage,omitempty"`
	//Network configuration, used for SSO and external access
	Network *network.Network `json:"network,omitempty" yaml:"network,omitempty"`
	//Authorization and Authentication configuration for SSO
	Auth *auth.Auth `json:"auth,omitempty" yaml:"auth,omitempty"`
	//List of plugins which get added to the grafana instance
	Plugins []string `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	//NodeSelector for deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run grafana on nodes
	Tolerations k8s.Tolerations `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	//Resource requirements
	Resources *k8s.Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
}

func (m *Monitoring) InitSecrets() {
	if m.Admin == nil {
		m.Admin = &admin.Admin{}
	}
	m.Admin.InitSecrets()

	if m.Auth == nil {
		m.Auth = &auth.Auth{}
	}
	m.Auth.InitSecrets()
}

func (m *Monitoring) IsZero() bool {
	if !m.Deploy &&
		(m.Admin == nil || m.Admin.IsZero()) &&
		(m.Auth == nil || m.Auth.IsZero()) &&
		m.Datasources == nil &&
		m.DashboardProviders == nil &&
		m.Storage == nil &&
		m.Network == nil &&
		m.Plugins == nil &&
		m.NodeSelector == nil &&
		m.Tolerations == nil &&
		m.Resources == nil {
		return true
	}

	return false
}

type Datasource struct {
	//Name of the datasource
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Type of the datasource (for example prometheus)
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	//URL to the datasource
	Url string `json:"url,omitempty" yaml:"url,omitempty"`
	//Access defintion of the datasource
	Access string `json:"access,omitempty" yaml:"access,omitempty"`
	//Boolean if datasource should be used as default
	IsDefault bool `json:"isDefault,omitempty" yaml:"isDefault,omitempty"`
}

type Provider struct {
	//ConfigMaps in which the dashboards are stored
	ConfigMaps []string `json:"configMaps,omitempty" yaml:"configMaps,omitempty"`
	//Local folder in which the dashboards are mounted
	Folder string `json:"folder,omitempty" yaml:"folder,omitempty"`
}
