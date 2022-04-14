package benchmarks

import (
	"context"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	lapi "github.com/filecoin-project/lotus/api"
)

// run chain head over and over
func HeadBench(ctx context.Context, delay time.Duration, api lapi.Gateway, r Results) error {
	var method = func(ctx context.Context) error {
		_, err := api.ChainHead(ctx)
		return err
	}
	return recordCount(ctx, delay, method, r)
}

// get the genesis tipset over and over
func GetGenesisBench(ctx context.Context, delay time.Duration, api lapi.Gateway, r Results) error {
	var method = func(ctx context.Context) error {
		_, err := api.ChainGetGenesis(ctx)
		return err
	}
	return recordCount(ctx, delay, method, r)
}

// this tends to walk back from head, fetching tipsets.
// Sometimes multiple goroutines might fetch the same tipset if ts is set to the same value.
// The goal is just to work out the ChainGetTipset method and reduce any possible caching
// by frequently changing the tipset being requested.
// The cache avoidance doesn't work as well when there are multple concurrent routines, since they
// all walk backward starting from head.
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

// Work out StateMinerInfo
// A list of miners is fetched at the start of the benchmark, and then they are looped over in a random
// order. If multiple concurrent routines are working, they'll work the miners in a different order
// to minimize caching.
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

		// Don't lookup the same miners as other routines or past runs.
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(miners), func(i, j int) {
			miners[i], miners[j] = miners[j], miners[i]
		})

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
