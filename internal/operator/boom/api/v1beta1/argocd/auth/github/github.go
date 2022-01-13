package github

import (
	secret2 "github.com/caos/orbos/pkg/secret"
)

type Connector struct {
	//Internal id of the github provider
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	//Internal name of the github provider
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Configuration for the github provider
	Config *Config `json:"config,omitempty" yaml:"config,omitempty"`
}

func (c *Connector) IsZero() bool {
	if c.ID == "" &&
		c.Name == "" &&
		(c.Config == nil || c.Config.IsZero()) {
		return true
	}

	return false
}

func (c *Config) IsZero() bool {
	if (c.ClientID == nil || c.ClientID.IsZero()) &&
		(c.ClientSecret == nil || c.ClientSecret.IsZero()) &&
		c.ExistingClientIDSecret == nil &&
		c.ExistingClientSecretSecret == nil &&
		c.Orgs == nil &&
		!c.LoadAllGroups &&
		c.TeamNameField == "" &&
		!c.UseLoginAsID {
		return true
	}
	return false
}

type Config struct {
	ClientID *secret2.Secret `json:"clientID,omitempty" yaml:"clientID,omitempty"`
	//Existing secret with the clientID
	ExistingClientIDSecret *secret2.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret           *secret2.Secret   `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
	//Existing secret with the clientSecret
	ExistingClientSecretSecret *secret2.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	//Required membership to organization in github
	Orgs []*Org `json:"orgs,omitempty" yaml:"orgs,omitempty"`
	//Flag which indicates that all user groups and teams should be loaded
	LoadAllGroups bool `json:"loadAllGroups,omitempty" yaml:"loadAllGroups,omitempty"`
	//Optional choice between 'name' (default), 'slug', or 'both'
	TeamNameField string `json:"teamNameField,omitempty" yaml:"teamNameField,omitempty"`
	//Flag which will switch from using the internal GitHub id to the users handle (@mention) as the user id
	UseLoginAsID bool `json:"useLoginAsID,omitempty" yaml:"useLoginAsID,omitempty"`
}

type Org struct {
	//Name of the organization
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Name of the required team in the organization
	Teams []string `json:"teams,omitempty" yaml:"teams,omitempty"`
}
