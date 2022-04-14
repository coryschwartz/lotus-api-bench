package benchmarks

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	lapi "github.com/filecoin-project/lotus/api"
)

func HeadBench(ctx context.Context, delay time.Duration, api lapi.Gateway, r Results) error {
	var method = func(ctx context.Context) error {
		_, err := api.ChainHead(ctx)
		return err
	}
	return recordCount(ctx, delay, method, r)
}

func GetGenesisBench(ctx context.Context, delay time.Duration, api lapi.Gateway, r Results) error {
	var method = func(ctx context.Context) error {
		_, err := api.ChainGetGenesis(ctx)
		return err
	}
	return recordCount(ctx, delay, method, r)
}

func WalkBack(ctx context.Context, delay time.Duration, api lapi.Gateway, r Results) error {
	ts, err := api.ChainHead(ctx)
	if err != nil {
		return fmt.Errorf("could not get head during walk: %w", err)
	}
	method := func(ctx context.Context) error {
		next, err := api.ChainGetTipSet(ctx, ts.Parents())
		if err != nil {
			return err
		}
		ts = next
		return nil
	}
	return recordCount(ctx, delay, method, r)
}

func InspectMiners(ctx context.Context, delay time.Duration, api lapi.Gateway, r Results) error {
	var minerlist int64
	var minerlistsecs int64
	var minerinfo int64
	var errcount int64

	ts, err := api.ChainHead(ctx)
	if err != nil {
		return fmt.Errorf("could not get head during miner inspection: %w", err)
	}
	for {
		start := time.Now()
		miners, err := api.StateListMiners(ctx, ts.Key())
		end := time.Now()
		if err != nil {
			return err
		} else {
			atomic.AddInt64(&minerlist, 1)
			atomic.AddInt64(&minerlistsecs, int64(end.Second()-start.Second()))
		}

		// loop over every miner and get info
		ticker := time.NewTicker(delay)
		for _, miner := range miners {
			select {
			case <-ticker.C:
				// The context is canceled when the bench is over.
				if ctx.Err() != nil {
					r["minerlist"] = minerlist
					r["minerlistsecs"] = minerlistsecs
					r["minerinfo"] = minerinfo
					r["errcount"] = errcount
					return nil
				}
				go func() {
					_, err := api.StateMinerInfo(ctx, miner, ts.Key())
					if err != nil {
						atomic.AddInt64(&errcount, 1)
						return
					}
					atomic.AddInt64(&minerinfo, 1)
				}()
			}
		}
		ticker.Stop()

		next, err := api.ChainGetTipSet(ctx, ts.Parents())
		if err != nil {
			atomic.AddInt64(&errcount, 1)
			continue
		} else {
			ts = next
		}

	}
}
