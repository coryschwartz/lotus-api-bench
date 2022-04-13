package benchmarks

import (
	"context"
	"fmt"
	"time"
)

func recordCount(ctx context.Context, delay time.Duration, method func(context.Context) error, r Results) error {
	var loopcount = 0
	var errcount = 0
	errs := make(chan error, 1)
	ticker := time.NewTicker(delay)
	go func() {
		for {
			select {
			case <-ticker.C:
				loopcount++
				// The context is canceled when the bench is over.
				if ctx.Err() != nil {
					r["loopcount"] = float64(loopcount)
					r["errcount"] = float64(errcount)
					errs <- nil
					return
				}
				err := method(ctx)
				if err != nil {
					errcount++
					fmt.Println(err)
				}
			}
		}
	}()
	return <-errs
}
