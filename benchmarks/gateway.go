package benchmarks

import (
	"context"
	"time"

	lapi "github.com/filecoin-project/lotus/api"
)

func GwHeadBench(ctx context.Context, delay time.Duration, api lapi.Gateway, r Results) error {
	var method = func(ctx context.Context) error {
		_, err := api.ChainHead(ctx)
		return err
	}
	return recordCount(ctx, delay, method, r)
}
