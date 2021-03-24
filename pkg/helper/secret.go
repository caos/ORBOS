package helper

import (
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/secret"
)

func IsExistentSecret(secret *secret.Secret, existent *secret.Existing) bool {
	if (secret == nil || secret.Value == "") && existent != nil && (existent.Name != "" && existent.Key != "") {
		return true
	}
	return false
}

func IsExistentClientSecret(existent *secret.ExistingIDSecret) bool {
	if existent != nil && (existent.Name != "" && existent.IDKey != "" && existent.SecretKey != "") {
		return true
	}
	return false
}

func IsCrdSecret(secret *secret.Secret, existent *secret.Existing) bool {
	if (secret != nil && secret.Value != "") && (existent == nil || (existent.Name == "" || existent.Key == "")) {
		return true
	}
	return false
}

func GetSecretValueOnlyIncluster(secret *secret.Secret, existing *secret.Existing) (string, error) {
	if IsExistentSecret(secret, existing) {
		secret, err := clientgo.GetSecret(existing.Name, "caos-system")
		if err != nil {
			return "", err
		}

		return string(secret.Data[existing.Key]), nil
	} else if IsCrdSecret(secret, existing) {
		return secret.Value, nil
	}

	return "", nil
}

func GetSecretValue(k8sClient kubernetes.ClientInt, secret *secret.Secret, existing *secret.Existing) (string, error) {
	if IsExistentSecret(secret, existing) {
		secret, err := k8sClient.GetSecret("caos-system", existing.Name)
		if err != nil {
			return "", err
		}

		return string(secret.Data[existing.Key]), nil
	} else if IsCrdSecret(secret, existing) {
		return secret.Value, nil
	}

	return "", nil
}
