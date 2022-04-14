package benchmarks

import (
	"context"
	"sync/atomic"
	"time"
)

func recordCount(ctx context.Context, delay time.Duration, method func(context.Context) error, r Results) error {
	var loopcount int64
	var errcount int64
	ticker := time.NewTicker(delay)
	for {
		select {
		case <-ticker.C:
			// The context is canceled when the bench is over.
			if ctx.Err() != nil {
				r["loopcount"] = loopcount
				r["errcount"] = errcount
				return nil
			}
			go func() {
				err := method(ctx)
				atomic.AddInt64(&loopcount, 1)
				if err != nil {
					atomic.AddInt64(&errcount, 1)
				}
			}()
		}
	}
}
