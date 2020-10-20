package legacycf

import (
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/legacycf/config"
	"github.com/caos/orbos/internal/secret"
)

func getSecretsMap(desiredKind *Desired) map[string]*secret.Secret {
	secrets := map[string]*secret.Secret{}
	if desiredKind.Spec == nil {
		desiredKind.Spec = &config.ExternalConfig{}
	}

	if desiredKind.Spec.Credentials == nil {
		desiredKind.Spec.Credentials = &config.Credentials{}
	}

	if desiredKind.Spec.Credentials.User == nil {
		desiredKind.Spec.Credentials.User = &secret.Secret{}
	}

	if desiredKind.Spec.Credentials.APIKey == nil {
		desiredKind.Spec.Credentials.APIKey = &secret.Secret{}
	}

	if desiredKind.Spec.Credentials.UserServiceKey == nil {
		desiredKind.Spec.Credentials.UserServiceKey = &secret.Secret{}
	}

	secrets["credentials.user"] = desiredKind.Spec.Credentials.User
	secrets["credentials.apikey"] = desiredKind.Spec.Credentials.APIKey
	secrets["credentials.userservicekey"] = desiredKind.Spec.Credentials.UserServiceKey

	return secrets
}
