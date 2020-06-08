package info

import "github.com/caos/orbos/internal/operator/boom/name"

const (
	applicationName name.Application = "prometheus-operator"
	namespace       string           = "caos-system"
)

func GetName() name.Application {
	return applicationName
}

func GetNamespace() string {
	return namespace
}
