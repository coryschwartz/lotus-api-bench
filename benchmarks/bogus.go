package benchmarks

import (
	"context"
	"time"

	lapi "github.com/filecoin-project/lotus/api"
)

func BogusBenchFunc(ctx context.Context, delay time.Duration, api lapi.FullNode, r Results) error {
	return nil
}
