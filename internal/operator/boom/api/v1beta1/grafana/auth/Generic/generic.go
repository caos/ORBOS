package generic

import (
	secret2 "github.com/caos/orbos/pkg/secret"
)

type Auth struct {
	ClientID *secret2.Secret `json:"clientID,omitempty" yaml:"clientID,omitempty"`
	//Existing secret with the clientID
	ExistingClientIDSecret *secret2.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret           *secret2.Secret   `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
	//Existing secret with the clientSecret
	ExistingClientSecretSecret *secret2.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	//Used scopes for the OAuth-flow
	Scopes []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	//Auth-endpoint
	AuthURL string `json:"authURL,omitempty" yaml:"authURL,omitempty"`
	//Token-endpoint
	TokenURL string `json:"tokenURL,omitempty" yaml:"tokenURL,omitempty"`
	//Userinfo-endpoint
	APIURL string `json:"apiURL,omitempty" yaml:"apiURL,omitempty"`
	//Domains allowed to login
	AllowedDomains []string `json:"allowedDomains,omitempty" yaml:"allowedDomains,omitempty"`
}

func (a *Auth) IsZero() bool {
	if (a.ClientID == nil || a.ClientID.IsZero()) &&
		(a.ClientSecret == nil || a.ClientSecret.IsZero()) &&
		a.ExistingClientIDSecret == nil &&
		a.ExistingClientSecretSecret == nil &&
		a.Scopes == nil &&
		a.AuthURL == "" &&
		a.TokenURL == "" &&
		a.APIURL == "" &&
		a.AllowedDomains == nil {
		return true
	}

	return false
}
