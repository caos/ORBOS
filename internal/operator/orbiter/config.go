package orbiter

import (
	"github.com/caos/orbos/internal/ingestion"
	"github.com/caos/orbos/pkg/git"
	"github.com/caos/orbos/pkg/orb"
)

type Config struct {
	OrbiterCommit string
	GitClient     *git.Client
	Adapt         AdaptFunc
	FinishedChan  chan struct{}
	PushEvents    func(events []*ingestion.EventRequest) error
	OrbConfig     orb.Orb
}
