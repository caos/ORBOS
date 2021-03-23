package auth

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling/auth/github"
	"github.com/caos/orbos/pkg/helper"
)

type githubConnector struct {
	ClientID      string `yaml:"clientID,omitempty"`
	ClientSecret  string `yaml:"clientSecret,omitempty"`
	RedirectURI   string `yaml:"redirectURI,omitempty"`
	Orgs          []*org `yaml:"orgs,omitempty"`
	LoadAllGroups bool   `yaml:"loadAllGroups,omitempty"`
	TeamNameField string `yaml:"teamNameField,omitempty"`
	UseLoginAsID  bool   `yaml:"useLoginAsID,omitempty"`
}
type org struct {
	Name  string   `yaml:"name,omitempty"`
	Teams []string `yaml:"teams,omitempty"`
}

func getGithub(spec *github.Connector, redirect string) (interface{}, error) {
	clientID, err := helper.GetSecretValueOnlyIncluster(spec.Config.ClientID, spec.Config.ExistingClientIDSecret)
	if err != nil {
		return nil, err
	}

	clientSecret, err := helper.GetSecretValueOnlyIncluster(spec.Config.ClientSecret, spec.Config.ExistingClientSecretSecret)
	if err != nil {
		return nil, err
	}

	if clientID == "" || clientSecret == "" {
		return nil, nil
	}

	var orgs []*org
	if len(spec.Config.Orgs) > 0 {
		orgs = make([]*org, len(spec.Config.Orgs))
		for k, v := range spec.Config.Orgs {
			orgs[k] = &org{
				Name:  v.Name,
				Teams: v.Teams,
			}
		}
	}

	github := &githubConnector{
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		RedirectURI:   redirect,
		Orgs:          orgs,
		LoadAllGroups: spec.Config.LoadAllGroups,
		TeamNameField: spec.Config.TeamNameField,
		UseLoginAsID:  spec.Config.UseLoginAsID,
	}

	return github, nil
}
