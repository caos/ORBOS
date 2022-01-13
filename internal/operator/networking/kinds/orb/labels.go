package orb

import "github.com/caos/orbos/pkg/labels"

func mustDatabaseOperator(binaryVersion *string) *labels.Operator {

	version := "unknown"
	if binaryVersion != nil {
		version = *binaryVersion
	}

	return labels.MustForOperator("ORBOS", "networking.caos.ch", version)
}
