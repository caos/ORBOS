package labels

import (
	"errors"

	"gopkg.in/yaml.v3"
)

var _ IDLabels = (*Name)(nil)

type Name struct {
	model InternalName
	base  *Component
}

type InternalNameProp struct {
	Name string `yaml:"app.kubernetes.io/name,omitempty"`
}

type InternalName struct {
	InternalNameProp  `yaml:",inline"`
	InternalComponent `yaml:",inline"`
}

func ForName(l *Component, name string) (*Name, error) {
	if name == "" {
		return nil, errors.New("name must not be nil")
	}
	return newName(l, name), nil
}

func newName(c *Component, name string) *Name {
	return &Name{
		base: c,
		model: InternalName{
			InternalNameProp:  InternalNameProp{Name: name},
			InternalComponent: c.model,
		},
	}
}

func NameFrom(arbitrary map[string]string) (*Name, error) {
	intermediate, err := yaml.Marshal(arbitrary)
	if err != nil {
		panic(err)
	}
	n := &Name{}
	return n, yaml.Unmarshal(intermediate, n)
}

func MustForName(l *Component, name string) *Name {
	n, err := ForName(l, name)
	if err != nil {
		panic(err)
	}
	return n
}

func (l *Name) Name() string {
	return l.model.Name
}

/*
func (l *Name) Major() int8 {
	return l.base.Major()
}
*/

func (l *Name) Equal(r comparable) bool {
	if right, ok := r.(*Name); ok {
		return l.model == right.model
	}
	return false
}

func (l *Name) MarshalYAML() (interface{}, error) {
	return l.model, nil
}

func (l *Name) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&l.model); err != nil {
		return err
	}
	l.base = &Component{}
	return node.Decode(l.base)
}
