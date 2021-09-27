package exit

import (
	"os"

	"github.com/caos/orbos/v5/internal/operator/nodeagent"
	"github.com/caos/orbos/v5/internal/operator/nodeagent/rebooter"
)

func New() nodeagent.Rebooter {
	return rebooter.Func(func() error {
		os.Exit(0)
		return nil
	})
}
