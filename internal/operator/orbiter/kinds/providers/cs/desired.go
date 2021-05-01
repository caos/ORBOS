package cs

import (
	"fmt"

	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

type Desired struct {
	Common        *tree.Common `yaml:",inline"`
	Spec          Spec
	Loadbalancing *tree.Tree
}

type Pool struct {
	Flavor       string
	Zone         string
	VolumeSizeGB int
}

func (p Pool) validate() error {
	return nil
}

type Spec struct {
	Verbose             bool
	APIToken            *secret.Secret `yaml:",omitempty"`
	Pools               map[string]*Pool
	SSHKey              *SSHKey
	RebootRequired      []string
	ReplacementRequired []string
}

type SSHKey struct {
	Private *secret.Secret `yaml:",omitempty"`
	Public  *secret.Secret `yaml:",omitempty"`
}

func (d Desired) validateAdapt() error {
	if d.Loadbalancing == nil {
		return errors.New("no loadbalancing configured")
	}
	if len(d.Spec.Pools) == 0 {
		return errors.New("no pools configured")
	}
	for poolName, pool := range d.Spec.Pools {
		if err := pool.validate(); err != nil {
			return fmt.Errorf("configuring pool %s failed: %w", poolName, err)
		}
	}
	return nil
}

func (d Desired) validateAPIToken() error {
	if d.Spec.APIToken == nil ||
		d.Spec.APIToken.Value == "" {
		return errors.New("apitoken missing... please provide a cloudscale api token using orbctl writesecret command")
	}
	return nil
}

func (d Desired) validateQuery() error {

	if err := d.validateAPIToken(); err != nil {
		return err
	}

	if d.Spec.SSHKey.Private == nil ||
		d.Spec.SSHKey.Private.Value == "" ||
		d.Spec.SSHKey.Public == nil ||
		d.Spec.SSHKey.Public.Value == "" {
		return errors.New("ssh key missing... please initialize your orb using orbctl configure command")
	}

	return nil
}

func parseDesired(desiredTree *tree.Tree) (*Desired, error) {
	desiredKind := &Desired{
		Common: desiredTree.Common,
		Spec:   Spec{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}
